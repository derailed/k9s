package ui

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// App represents an application.
type App struct {
	*tview.Application
	Configurator

	Main    *Pages
	flash   *model.Flash
	actions KeyActions
	views   map[string]tview.Primitive
	cmdBuff *CmdBuff
}

// NewApp returns a new app.
func NewApp(context string) *App {
	a := App{
		Application: tview.NewApplication(),
		actions:     make(KeyActions),
		Main:        NewPages(),
		flash:       model.NewFlash(model.DefaultFlashDelay),
		cmdBuff:     NewCmdBuff(':', CommandBuff),
	}
	a.ReloadStyles(context)

	a.views = map[string]tview.Primitive{
		"menu":   NewMenu(a.Styles),
		"logo":   NewLogo(a.Styles),
		"cmd":    NewCommand(a.Styles),
		"crumbs": NewCrumbs(a.Styles),
	}

	return &a
}

// Init initializes the application.
func (a *App) Init() {
	a.bindKeys()
	a.cmdBuff.AddListener(a.Cmd())
	a.Styles.AddListener(a)
	a.CmdBuff().AddListener(a)

	a.SetRoot(a.Main, true)
}

// BufferChanged indicates the buffer was changed.
func (a *App) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (a *App) BufferActive(state bool, _ BufferKind) {
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		return
	}

	if state && flex.ItemAt(1) != a.Cmd() {
		flex.AddItemAtIndex(1, a.Cmd(), 3, 1, false)
	} else if !state && flex.ItemAt(1) == a.Cmd() {
		flex.RemoveItemAtIndex(1)
	}
	a.Draw()
}

// StylesChanged notifies the skin changed.
func (a *App) StylesChanged(s *config.Styles) {
	a.Main.SetBackgroundColor(s.BgColor())
	if f, ok := a.Main.GetPrimitive("main").(*tview.Flex); ok {
		f.SetBackgroundColor(s.BgColor())
		if h, ok := f.ItemAt(0).(*tview.Flex); ok {
			h.SetBackgroundColor(s.BgColor())
		} else {
			log.Error().Msgf("Header not found")
		}
	} else {
		log.Error().Msgf("Main not found")
	}
}

// ReloadStyles reloads skin file.
func (a *App) ReloadStyles(context string) {
	a.RefreshStyles(context)
}

// Conn returns an api server connection.
func (a *App) Conn() client.Connection {
	return a.Config.GetConnection()
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
		tcell.KeyCtrlU:      NewSharedKeyAction("Clear Filter", a.clearCmd, false),
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

// HasAction checks if key matches a registered binding.
func (a *App) HasAction(key tcell.Key) (KeyAction, bool) {
	act, ok := a.actions[key]
	return act, ok
}

// GetActions returns a collection of actiona.
func (a *App) GetActions() KeyActions {
	return a.actions
}

// AddActions returns the application actiona.
func (a *App) AddActions(aa KeyActions) {
	for k, v := range aa {
		a.actions[k] = v
	}
}

// Views return the application root viewa.
func (a *App) Views() map[string]tview.Primitive {
	return a.views
}

func (a *App) clearCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.CmdBuff().IsActive() {
		return evt
	}
	a.CmdBuff().Clear()

	return nil
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

// View Accessora...

// Crumbs return app crumba.
func (a *App) Crumbs() *Crumbs {
	return a.views["crumbs"].(*Crumbs)
}

// Logo return the app logo.
func (a *App) Logo() *Logo {
	return a.views["logo"].(*Logo)
}

// Cmd returns app cmd.
func (a *App) Cmd() *Command {
	return a.views["cmd"].(*Command)
}

// Menu returns app menu.
func (a *App) Menu() *Menu {
	return a.views["menu"].(*Menu)
}

// Flash returns a flash model.
func (a *App) Flash() *model.Flash {
	return a.flash
}

// ----------------------------------------------------------------------------
// Helpers...

// AsKey converts rune to keyboard key.,
func AsKey(evt *tcell.EventKey) tcell.Key {
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
