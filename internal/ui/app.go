// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"os"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

// App represents an application.
type App struct {
	*tview.Application
	Configurator

	Main    *Pages
	flash   *model.Flash
	actions *KeyActions
	views   map[string]tview.Primitive
	cmdBuff *model.FishBuff
	running bool
	mx      sync.RWMutex
}

// NewApp returns a new app.
func NewApp(cfg *config.Config, context string) *App {
	a := App{
		Application:  tview.NewApplication(),
		actions:      NewKeyActions(),
		Configurator: Configurator{Config: cfg, Styles: config.NewStyles()},
		Main:         NewPages(),
		flash:        model.NewFlash(model.DefaultFlashDelay),
		cmdBuff:      model.NewFishBuff(':', model.CommandBuffer),
	}

	a.views = map[string]tview.Primitive{
		"menu":   NewMenu(a.Styles),
		"logo":   NewLogo(a.Styles),
		"prompt": NewPrompt(&a, a.Config.K9s.UI.NoIcons, a.Styles),
		"crumbs": NewCrumbs(a.Styles),
	}

	return &a
}

// Init initializes the application.
func (a *App) Init() {
	a.bindKeys()
	a.Prompt().SetModel(a.cmdBuff)
	a.cmdBuff.AddListener(a)
	a.Styles.AddListener(a)

	a.SetRoot(a.Main, true).EnableMouse(a.Config.K9s.UI.EnableMouse)
}

// QueueUpdate queues up a ui action.
func (a *App) QueueUpdate(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdate(f)
	}()
}

// QueueUpdateDraw queues up a ui action and redraw the ui.
func (a *App) QueueUpdateDraw(f func()) {
	if a.Application == nil {
		return
	}
	go func() {
		a.Application.QueueUpdateDraw(f)
	}()
}

// IsRunning checks if app is actually running.
func (a *App) IsRunning() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()
	return a.running
}

// SetRunning sets the app run state.
func (a *App) SetRunning(f bool) {
	a.mx.Lock()
	defer a.mx.Unlock()
	a.running = f
}

// BufferCompleted indicates input was accepted.
func (a *App) BufferCompleted(_, _ string) {}

// BufferChanged indicates the buffer was changed.
func (a *App) BufferChanged(_, _ string) {}

// BufferActive indicates the buff activity changed.
func (a *App) BufferActive(state bool, kind model.BufferKind) {
	flex, ok := a.Main.GetPrimitive("main").(*tview.Flex)
	if !ok {
		return
	}

	if state && flex.ItemAt(1) != a.Prompt() {
		flex.AddItemAtIndex(1, a.Prompt(), 3, 1, false)
	} else if !state && flex.ItemAt(1) == a.Prompt() {
		flex.RemoveItemAtIndex(1)
		a.SetFocus(flex)
	}
}

// SuggestionChanged notifies of update to command suggestions.
func (a *App) SuggestionChanged(ss []string) {}

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

// Conn returns an api server connection.
func (a *App) Conn() client.Connection {
	return a.Config.GetConnection()
}

func (a *App) bindKeys() {
	a.actions = NewKeyActionsFromMap(KeyMap{
		KeyColon:       NewKeyAction("Cmd", a.activateCmd, false),
		tcell.KeyCtrlR: NewKeyAction("Redraw", a.redrawCmd, false),
		tcell.KeyCtrlP: NewKeyAction("Persist", a.saveCmd, false),
		tcell.KeyCtrlU: NewSharedKeyAction("Clear Filter", a.clearCmd, false),
		tcell.KeyCtrlQ: NewSharedKeyAction("Clear Filter", a.clearCmd, false),
	})
}

// BailOut exits the application.
func (a *App) BailOut() {
	if err := a.Config.Save(true); err != nil {
		log.Error().Err(err).Msg("config save failed!")
	}

	a.Stop()
	os.Exit(0)
}

// ResetPrompt reset the prompt model and marks buffer as active.
func (a *App) ResetPrompt(m PromptModel) {
	m.ClearText(false)
	a.Prompt().SetModel(m)
	a.SetFocus(a.Prompt())
	m.SetActive(true)
}

// ResetCmd clear out user command.
func (a *App) ResetCmd() {
	a.cmdBuff.Reset()
}

func (a *App) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if err := a.Config.Save(true); err != nil {
		a.Flash().Err(err)
	}
	a.Flash().Info("current context config saved")

	return nil
}

// ActivateCmd toggle command mode.
func (a *App) ActivateCmd(b bool) {
	a.cmdBuff.SetActive(b)
}

// GetCmd retrieves user command.
func (a *App) GetCmd() string {
	return a.cmdBuff.GetText()
}

// CmdBuff returns the app cmd model.
func (a *App) CmdBuff() *model.FishBuff {
	return a.cmdBuff
}

// HasCmd check if cmd buffer is active and has a command.
func (a *App) HasCmd() bool {
	return a.cmdBuff.IsActive() && !a.cmdBuff.Empty()
}

// InCmdMode check if command mode is active.
func (a *App) InCmdMode() bool {
	return a.Prompt().InCmdMode()
}

// HasAction checks if key matches a registered binding.
func (a *App) HasAction(key tcell.Key) (KeyAction, bool) {
	return a.actions.Get(key)
}

// GetActions returns a collection of actions.
func (a *App) GetActions() *KeyActions {
	return a.actions
}

// AddActions returns the application actions.
func (a *App) AddActions(aa *KeyActions) {
	a.actions.Merge(aa)
}

// Views return the application root views.
func (a *App) Views() map[string]tview.Primitive {
	return a.views
}

func (a *App) clearCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !a.cmdBuff.IsActive() {
		return evt
	}
	a.cmdBuff.ClearText(true)

	return nil
}

func (a *App) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if a.InCmdMode() {
		return evt
	}
	a.ResetPrompt(a.cmdBuff)
	a.cmdBuff.ClearText(true)

	return nil
}

// RedrawCmd forces a redraw.
func (a *App) redrawCmd(evt *tcell.EventKey) *tcell.EventKey {
	a.QueueUpdateDraw(func() {})
	return evt
}

// View Accessors...

// Crumbs return app crumbs.
func (a *App) Crumbs() *Crumbs {
	return a.views["crumbs"].(*Crumbs)
}

// Logo return the app logo.
func (a *App) Logo() *Logo {
	return a.views["logo"].(*Logo)
}

// Prompt returns command prompt.
func (a *App) Prompt() *Prompt {
	return a.views["prompt"].(*Prompt)
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

// AsKey converts rune to keyboard key.
func AsKey(evt *tcell.EventKey) tcell.Key {
	if evt.Key() != tcell.KeyRune {
		return evt.Key()
	}
	key := tcell.Key(evt.Rune())
	if evt.Modifiers() == tcell.ModAlt {
		key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
	}
	return key
}
