package ui

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// App represents an application.
type App struct {
	*tview.Application
	Configurator

	Main *Pages
	Hint *model.Hint

	actions KeyActions

	views   map[string]tview.Primitive
	cmdBuff *CmdBuff
}

// NewApp returns a new app.
func NewApp() *App {
	a := App{
		Application: tview.NewApplication(),
		actions:     make(KeyActions),
		Main:        NewPages(),
		cmdBuff:     NewCmdBuff(':', CommandBuff),
		Hint:        model.NewHint(),
	}

	a.RefreshStyles()

	a.views = map[string]tview.Primitive{
		"menu":   NewMenu(a.Styles),
		"logo":   NewLogoView(a.Styles),
		"cmd":    NewCmdView(a.Styles),
		"flash":  NewFlashView(&a, "Initializing..."),
		"crumbs": NewCrumbs(a.Styles),
	}

	return &a
}

// Init initializes the application.
func (a *App) Init() {
	a.bindKeys()
	a.SetInputCapture(a.keyboard)
	a.cmdBuff.AddListener(a.Cmd())
	a.SetRoot(a.Main, true)

	a.Hint.AddListener(a.Menu())
}

// Conn returns an api server connection.
func (a *App) Conn() k8s.Connection {
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

// StatusReset reset log back to normal.
func (a *App) StatusReset() {
	a.Logo().Reset()
	a.Draw()
}

// View Accessora...

// Crumbs return app crumba.
func (a *App) Crumbs() *Crumbs {
	return a.views["crumbs"].(*Crumbs)
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
func (a *App) Menu() *Menu {
	return a.views["menu"].(*Menu)
}

// AsKey converts rune to keyboard key.,
func asKey(evt *tcell.EventKey) tcell.Key {
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
