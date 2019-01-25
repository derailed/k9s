package tview

import (
	"sync"

	"github.com/gdamore/tcell"
)

// The size of the event/update/redraw channels.
const queueSize = 100

// Application represents the top node of an application.
//
// It is not strictly required to use this class as none of the other classes
// depend on it. However, it provides useful tools to set up an application and
// plays nicely with all widgets.
type Application struct {
	sync.RWMutex

	// The application's screen.
	screen tcell.Screen

	// Indicates whether the application's screen is currently active. This is
	// false during suspended mode.
	running bool

	// The primitive which currently has the keyboard focus.
	focus Primitive

	// The root primitive to be seen on the screen.
	root Primitive

	// Whether or not the application resizes the root primitive.
	rootFullscreen bool

	// An optional capture function which receives a key event and returns the
	// event to be forwarded to the default input handler (nil if nothing should
	// be forwarded).
	inputCapture func(event *tcell.EventKey) *tcell.EventKey

	// Toggle mouseEvents
	mouseEnable bool

	// Option callback to track mouse events.
	mouseCapture func(event *tcell.EventMouse)

	// An optional callback function which is invoked just before the root
	// primitive is drawn.
	beforeDraw func(screen tcell.Screen) bool

	// An optional callback function which is invoked after the root primitive
	// was drawn.
	afterDraw func(screen tcell.Screen)

	// Used to send screen events from separate goroutine to main event loop
	events chan tcell.Event

	// Functions queued from goroutines, used to serialize updates to primitives.
	updates chan func()

	// Redraw requests.
	redraw chan struct{}

	// A channel which signals the end of the suspended mode.
	suspendToken chan struct{}
}

// NewApplication creates and returns a new application.
func NewApplication() *Application {
	return &Application{
		events:       make(chan tcell.Event, queueSize),
		updates:      make(chan func(), queueSize),
		redraw:       make(chan struct{}, queueSize),
		suspendToken: make(chan struct{}, 1),
	}
}

// EnableMouse turns on mouse events
func (a *Application) EnableMouse(b bool) {
	a.mouseEnable = b
}

// SetMouseCapture enable mouse event listener.
func (a *Application) SetMouseCapture(c func(evt *tcell.EventMouse)) {
	a.mouseCapture = c
}

// SetInputCapture sets a function which captures all key events before they are
// forwarded to the key event handler of the primitive which currently has
// focus. This function can then choose to forward that key event (or a
// different one) by returning it or stop the key event processing by returning
// nil.
//
// Note that this also affects the default event handling of the application
// itself: Such a handler can intercept the Ctrl-C event which closes the
// applicatoon.
func (a *Application) SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *Application {
	a.inputCapture = capture
	return a
}

// GetInputCapture returns the function installed with SetInputCapture() or nil
// if no such function has been installed.
func (a *Application) GetInputCapture() func(event *tcell.EventKey) *tcell.EventKey {
	return a.inputCapture
}

// SetScreen allows you to provide your own tcell.Screen object. For most
// applications, this is not needed and you should be familiar with
// tcell.Screen when using this function. Run() will call Init() and Fini() on
// the provided screen object.
//
// This function is typically called before calling Run(). Calling it while an
// application is running will switch the application to the new screen. Fini()
// will be called on the old screen and Init() on the new screen (errors
// returned by Init() will lead to a panic).
//
// Note that calling Suspend() will invoke Fini() on your screen object and it
// will not be restored when suspended mode ends. Instead, a new default screen
// object will be created.
func (a *Application) SetScreen(screen tcell.Screen) *Application {
	a.Lock()
	defer a.Unlock()
	if a.running {
		a.screen.Fini()
	}
	a.screen = screen
	if a.running {
		if err := a.screen.Init(); err != nil {
			panic(err)
		}
	}
	return a
}

// Run starts the application and thus the event loop. This function returns
// when Stop() was called.
func (a *Application) Run() error {
	var err error
	a.Lock()

	// Make a screen if there is none yet.
	if a.screen == nil {
		a.screen, err = tcell.NewScreen()
		if err != nil {
			a.Unlock()
			return err
		}
	}
	if err = a.screen.Init(); err != nil {
		a.Unlock()
		return err
	}
	a.running = true

	if a.mouseEnable {
		a.screen.EnableMouse()
	}

	// We catch panics to clean up because they mess up the terminal.
	defer func() {
		if p := recover(); p != nil {
			if a.screen != nil {
				a.screen.Fini()
			}
			a.running = false
			panic(p)
		}
	}()

	// Draw the screen for the first time.
	a.Unlock()
	a.draw()

	// Separate loop to wait for screen events.
	var wg sync.WaitGroup
	wg.Add(1)
	a.suspendToken <- struct{}{} // We need this to get started.
	go func() {
		defer wg.Done()
		for range a.suspendToken {
			for {
				a.RLock()
				screen := a.screen
				a.RUnlock()
				if screen == nil {
					// We have no screen. We might need to stop.
					break
				}

				// Wait for next event and queue it.
				event := screen.PollEvent()
				if event != nil {
					// Regular event. Queue.
					a.QueueEvent(event)
					continue
				}

				// A screen was finalized (event is nil).
				a.RLock()
				running := a.running
				a.RUnlock()
				if running {
					// The application was stopped. End the event loop.
					a.QueueEvent(nil)
					return
				}

				// We're in suspended mode (running is false). Pause and wait for new
				// token.
				break
			}
		}
	}()

	// Start event loop.
EventLoop:
	for {
		select {
		case event := <-a.events:
			if event == nil {
				break EventLoop
			}

			switch event := event.(type) {
			case *tcell.EventMouse:
				if a.mouseCapture != nil {
					a.mouseCapture(event)
				}
			case *tcell.EventKey:
				a.RLock()
				p := a.focus
				inputCapture := a.inputCapture
				a.RUnlock()

				// Intercept keys.
				if inputCapture != nil {
					event = inputCapture(event)
					if event == nil {
						break EventLoop // Don't forward event.
					}
				}

				// Ctrl-C closes the application.
				if event.Key() == tcell.KeyCtrlC {
					a.Stop()
				}

				// Pass other key events to the currently focused primitive.
				if p != nil {
					if handler := p.InputHandler(); handler != nil {
						handler(event, func(p Primitive) {
							a.SetFocus(p)
						})
						a.draw()
					}
				}
			case *tcell.EventResize:
				a.RLock()
				screen := a.screen
				a.RUnlock()
				screen.Clear()
				a.draw()
			}

		// If we have updates, now is the time to execute them.
		case updater := <-a.updates:
			updater()

			// If a redraw is requested, do it now.
		case <-a.redraw:
			a.draw()
		}
	}

	a.running = false
	close(a.suspendToken)
	wg.Wait()

	return nil
}

// Stop stops the application, causing Run() to return.
func (a *Application) Stop() {
	a.Lock()
	defer a.Unlock()
	screen := a.screen
	if screen == nil {
		return
	}
	a.screen = nil
	screen.Fini()
	// a.running is still true, the main loop will clean up.
}

// Suspend temporarily suspends the application by exiting terminal UI mode and
// invoking the provided function "f". When "f" returns, terminal UI mode is
// entered again and the application resumes.
//
// A return value of true indicates that the application was suspended and "f"
// was called. If false is returned, the application was already suspended,
// terminal UI mode was not exited, and "f" was not called.
func (a *Application) Suspend(f func()) bool {
	a.Lock()

	screen := a.screen
	if screen == nil {
		// Screen has not yet been initialized.
		a.Unlock()
		return false
	}

	// Enter suspended mode. Make a new screen here already so our event loop can
	// continue.
	a.screen = nil
	a.running = false
	screen.Fini()
	a.Unlock()

	// Wait for "f" to return.
	f()

	// Initialize our new screen and draw the contents.
	a.Lock()
	var err error
	a.screen, err = tcell.NewScreen()
	if err != nil {
		a.Unlock()
		panic(err)
	}
	if err = a.screen.Init(); err != nil {
		a.Unlock()
		panic(err)
	}
	a.running = true
	a.Unlock()
	a.draw()
	a.suspendToken <- struct{}{}
	// One key event will get lost, see https://github.com/gdamore/tcell/issues/194

	// Continue application loop.
	return true
}

// Draw refreshes the screen. It calls the Draw() function of the application's
// root primitive and then syncs the screen buffer.
func (a *Application) Draw() *Application {
	// We actually just queue this draw.
	a.redraw <- struct{}{}
	return a
}

// draw actually does what Draw() promises to do.
func (a *Application) draw() *Application {
	a.Lock()
	defer a.Unlock()

	screen := a.screen
	root := a.root
	fullscreen := a.rootFullscreen
	before := a.beforeDraw
	after := a.afterDraw

	// Maybe we're not ready yet or not anymore.
	if screen == nil || root == nil {
		return a
	}

	// Resize if requested.
	if fullscreen && root != nil {
		width, height := screen.Size()
		root.SetRect(0, 0, width, height)
	}

	// Call before handler if there is one.
	if before != nil {
		if before(screen) {
			screen.Show()
			return a
		}
	}

	// Draw all primitives.
	root.Draw(screen)

	// Call after handler if there is one.
	if after != nil {
		after(screen)
	}

	// Sync screen.
	screen.Show()

	return a
}

// SetBeforeDrawFunc installs a callback function which is invoked just before
// the root primitive is drawn during screen updates. If the function returns
// true, drawing will not continue, i.e. the root primitive will not be drawn
// (and an after-draw-handler will not be called).
//
// Note that the screen is not cleared by the application. To clear the screen,
// you may call screen.Clear().
//
// Provide nil to uninstall the callback function.
func (a *Application) SetBeforeDrawFunc(handler func(screen tcell.Screen) bool) *Application {
	a.beforeDraw = handler
	return a
}

// GetBeforeDrawFunc returns the callback function installed with
// SetBeforeDrawFunc() or nil if none has been installed.
func (a *Application) GetBeforeDrawFunc() func(screen tcell.Screen) bool {
	return a.beforeDraw
}

// SetAfterDrawFunc installs a callback function which is invoked after the root
// primitive was drawn during screen updates.
//
// Provide nil to uninstall the callback function.
func (a *Application) SetAfterDrawFunc(handler func(screen tcell.Screen)) *Application {
	a.afterDraw = handler
	return a
}

// GetAfterDrawFunc returns the callback function installed with
// SetAfterDrawFunc() or nil if none has been installed.
func (a *Application) GetAfterDrawFunc() func(screen tcell.Screen) {
	return a.afterDraw
}

// SetRoot sets the root primitive for this application. If "fullscreen" is set
// to true, the root primitive's position will be changed to fill the screen.
//
// This function must be called at least once or nothing will be displayed when
// the application starts.
//
// It also calls SetFocus() on the primitive.
func (a *Application) SetRoot(root Primitive, fullscreen bool) *Application {
	a.Lock()
	a.root = root
	a.rootFullscreen = fullscreen
	if a.screen != nil {
		a.screen.Clear()
	}
	a.Unlock()

	a.SetFocus(root)

	return a
}

// ResizeToFullScreen resizes the given primitive such that it fills the entire
// screen.
func (a *Application) ResizeToFullScreen(p Primitive) *Application {
	a.RLock()
	width, height := a.screen.Size()
	a.RUnlock()
	p.SetRect(0, 0, width, height)
	return a
}

// SetFocus sets the focus on a new primitive. All key events will be redirected
// to that primitive. Callers must ensure that the primitive will handle key
// events.
//
// Blur() will be called on the previously focused primitive. Focus() will be
// called on the new primitive.
func (a *Application) SetFocus(p Primitive) *Application {
	a.Lock()
	if a.focus != nil {
		a.focus.Blur()
	}
	a.focus = p
	if a.screen != nil {
		a.screen.HideCursor()
	}
	a.Unlock()
	if p != nil {
		p.Focus(func(p Primitive) {
			a.SetFocus(p)
		})
	}

	return a
}

// GetFocus returns the primitive which has the current focus. If none has it,
// nil is returned.
func (a *Application) GetFocus() Primitive {
	a.RLock()
	defer a.RUnlock()
	return a.focus
}

// QueueUpdate is used to synchronize access to primitives from non-main
// goroutines. The provided function will be executed as part of the event loop
// and thus will not cause race conditions with other such update functions or
// the Draw() function.
//
// Note that Draw() is not implicitly called after the execution of f as that
// may not be desirable. You can call Draw() from f if the screen should be
// refreshed after each update.
func (a *Application) QueueUpdate(f func()) *Application {
	a.updates <- f
	return a
}

// QueueEvent sends an event to the Application event loop.
//
// It is not recommended for event to be nil.
func (a *Application) QueueEvent(event tcell.Event) *Application {
	a.events <- event
	return a
}
