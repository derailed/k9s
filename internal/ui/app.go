package ui

import (
	"context"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// Igniter represents an initializable view.
type Igniter interface {
	tview.Primitive

	// Init initializes the view.
	Init(ctx context.Context, ns string)
}

type (
	keyHandler interface {
		keyboard(evt *tcell.EventKey) *tcell.EventKey
	}

	// ActionsFunc augments Keybindings.
	ActionsFunc func(KeyActions)

	// Configurator represents an application configurations.
	Configurator struct {
		HasSkins bool
		Config   *config.Config
		Styles   *config.Styles
		Bench    *config.Bench
	}

	// App represents an application.
	App struct {
		*tview.Application
		Configurator

		actions KeyActions
		pages   *tview.Pages
		content *tview.Pages
		views   map[string]tview.Primitive
		cmdBuff *CmdBuff
		hints   Hints
	}
)

// NewApp returns a new app.
func NewApp() *App {
	s := App{
		Application: tview.NewApplication(),
		actions:     make(KeyActions),
		pages:       tview.NewPages(),
		content:     tview.NewPages(),
		cmdBuff:     NewCmdBuff(':', CommandBuff),
	}

	s.RefreshStyles()

	s.views = map[string]tview.Primitive{
		"menu":   NewMenuView(s.Styles),
		"logo":   NewLogoView(s.Styles),
		"cmd":    NewCmdView(s.Styles),
		"crumbs": NewCrumbsView(s.Styles),
	}

	return &s
}

// Main returns main app frame.
func (a *App) Main() *tview.Pages {
	return a.pages
}

// Frame returns main app content frame.
func (a *App) Frame() *tview.Pages {
	return a.content
}

// Conn returns an api server connection.
func (a *App) Conn() k8s.Connection {
	return a.Config.GetConnection()
}

// Init initializes the application.
func (a *App) Init() {
	a.bindKeys()
	a.SetInputCapture(a.keyboard)
	a.cmdBuff.AddListener(a.Cmd())
	a.SetRoot(a.pages, true)
}

func (a *App) bindKeys() {
	a.actions = KeyActions{
		KeyColon:            NewKeyAction("Cmd", a.activateCmd, false),
		tcell.KeyCtrlR:      NewKeyAction("Redraw", a.redrawCmd, false),
		tcell.KeyCtrlC:      NewKeyAction("Quit", a.quitCmd, false),
		tcell.KeyEscape:     NewKeyAction("Escape", a.escapeCmd, false),
		tcell.KeyBackspace2: NewKeyAction("Erase", a.eraseCmd, false),
		tcell.KeyBackspace:  NewKeyAction("Erase", a.eraseCmd, false),
		tcell.KeyDelete:     NewKeyAction("Erase", a.eraseCmd, false),
	}
}

// BailOut exists the application.
func (a *App) BailOut() {
	a.Stop()
}

// ResetCmd clear out user command.
func (a *App) ResetCmd() {
	a.cmdBuff.Reset()
}

// ActivateCmd toggle command mode.
func (a *App) ActivateCmd(b bool) {
	a.cmdBuff.SetActive(b)
}

// GetCmd retrieves user command.
func (a *App) GetCmd() string {
	return a.cmdBuff.String()
}

// CmdBuff returns a cmd buffer.
func (a *App) CmdBuff() *CmdBuff {
	return a.cmdBuff
}

// HasCmd check if cmd buffer is active and has a command.
func (a *App) HasCmd() bool {
	return a.cmdBuff.IsActive() && !a.cmdBuff.Empty()
}

func (a *App) quitCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}
	a.BailOut()

	return nil
}

// InCmdMode check if command mode is active.
func (a *App) InCmdMode() bool {
	return a.Cmd().InCmdMode()
}

// GetActions returns a collection of actions.
func (a *App) GetActions() KeyActions {
	return a.actions
}

// AddActions returns the application actions.
func (a *App) AddActions(aa KeyActions) {
	for k, v := range aa {
		a.actions[k] = v
	}
}

// Views return the application root views.
func (a *App) Views() map[string]tview.Primitive {
	return a.views
}

func (a *App) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if a.cmdBuff.IsActive() && evt.Modifiers() == tcell.ModNone {
			a.cmdBuff.Add(evt.Rune())
			return nil
		}
		key = asKey(evt)
	}

	if a, ok := a.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

func (a *App) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}
	a.cmdBuff.SetActive(true)
	a.cmdBuff.Clear()

	return nil
}

// EraseCmd removes the last char from a command.
func (a *App) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.cmdBuff.IsActive() {
		a.cmdBuff.Delete()
		return nil
	}
	return evt
}

// EscapeCmd dismiss cmd mode.
func (a *App) escapeCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.cmdBuff.IsActive() {
		a.cmdBuff.Reset()
	}
	return evt
}

// RedrawCmd forces a redraw.
func (a *App) redrawCmd(evt *tcell.EventKey) *tcell.EventKey {
	a.Draw()
	return evt
}

// ActiveView returns the currently active view.
func (a *App) ActiveView() Igniter {
	return a.content.GetPrimitive("main").(Igniter)
}

// SetHints updates menu hints.
func (a *App) SetHints(h Hints) {
	a.hints = h
	a.views["menu"].(*MenuView).HydrateMenu(h)
}

// GetHints retrieves the currently active hints.
func (a *App) GetHints() Hints {
	return a.hints
}

// StatusReset reset log back to normal.
func (a *App) StatusReset() {
	a.Logo().Reset()
	a.Draw()
}

// View Accessors...

// Crumbs return app crumbs.
func (a *App) Crumbs() *CrumbsView {
	return a.views["crumbs"].(*CrumbsView)
}

// Logo return the app logo.
func (a *App) Logo() *LogoView {
	return a.views["logo"].(*LogoView)
}

// Flash returns app flash.
func (a *App) Flash() *FlashView {
	return a.views["flash"].(*FlashView)
}

// Cmd returns app cmd.
func (a *App) Cmd() *CmdView {
	return a.views["cmd"].(*CmdView)
}

// Menu returns app menu.
func (a *App) Menu() *MenuView {
	return a.views["menu"].(*MenuView)
}

// AsKey converts rune to keyboard key.,
func asKey(evt *tcell.EventKey) tcell.Key {
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
