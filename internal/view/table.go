// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"path/filepath"
	"strings"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// Table represents a table viewer.
type Table struct {
	*ui.Table

	app        *App
	enterFn    EnterFunc
	envFn      EnvFunc
	bindKeysFn []BindKeysFunc
}

// NewTable returns a new viewer.
func NewTable(gvr client.GVR) *Table {
	t := Table{
		Table: ui.NewTable(gvr),
	}
	t.envFn = t.defaultEnv

	return &t
}

// Init initializes the component.
func (t *Table) Init(ctx context.Context) (err error) {
	if t.app, err = extractApp(ctx); err != nil {
		return err
	}
	if t.app.Conn() != nil {
		ctx = context.WithValue(ctx, internal.KeyHasMetrics, t.app.Conn().HasMetrics())
	}
	t.app.CustomView = config.NewCustomView()
	ctx = context.WithValue(ctx, internal.KeyStyles, t.app.Styles)
	ctx = context.WithValue(ctx, internal.KeyViewConfig, t.app.CustomView)
	t.Table.Init(ctx)
	if !t.app.Config.K9s.UI.Reactive {
		if err := t.app.RefreshCustomViews(); err != nil {
			log.Warn().Err(err).Msg("CustomViews load failed")
			t.app.Logo().Warn("Views load failed!")
		}
	}
	t.SetInputCapture(t.keyboard)
	t.bindKeys()
	t.GetModel().SetRefreshRate(time.Duration(t.app.Config.K9s.GetRefreshRate()) * time.Second)
	t.CmdBuff().AddListener(t)

	return nil
}

// HeaderIndex returns index of a given column or false if not found.
func (t *Table) HeaderIndex(colName string) (int, bool) {
	for i := 0; i < t.GetColumnCount(); i++ {
		h := t.GetCell(0, i)
		if h == nil {
			continue
		}
		s := h.Text
		if idx := strings.Index(s, "["); idx > 0 {
			s = s[:idx]
		}
		if s == colName {
			return i, true
		}
	}
	return 0, false
}

// SendKey sends a keyboard event (testing only!).
func (t *Table) SendKey(evt *tcell.EventKey) {
	t.app.Prompt().SendKey(evt)
}

func (t *Table) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyUp || key == tcell.KeyDown {
		return evt
	}

	if a, ok := t.Actions().Get(ui.AsKey(evt)); ok && !t.app.Content.IsTopDialog() {
		return a.Action(evt)
	}

	return evt
}

// Name returns the table name.
func (t *Table) Name() string { return t.GVR().R() }

// AddBindKeysFn adds additional key bindings.
func (t *Table) AddBindKeysFn(f BindKeysFunc) {
	t.bindKeysFn = append(t.bindKeysFn, f)
}

// SetEnvFn sets a function to pull viewer env vars for plugins.
func (t *Table) SetEnvFn(f EnvFunc) { t.envFn = f }

// EnvFn returns an plugin env function if available.
func (t *Table) EnvFn() EnvFunc {
	return t.envFn
}

func (t *Table) defaultEnv() Env {
	path := t.GetSelectedItem()
	row := t.GetSelectedRow(path)
	env := defaultEnv(t.app.Conn().Config(), path, t.GetModel().Peek().Header(), row)
	env["FILTER"] = t.CmdBuff().GetText()
	if env["FILTER"] == "" {
		env["NAMESPACE"], env["FILTER"] = client.Namespaced(path)
	}
	env["RESOURCE_GROUP"] = t.GVR().G()
	env["RESOURCE_VERSION"] = t.GVR().V()
	env["RESOURCE_NAME"] = t.GVR().R()

	return env
}

// App returns the current app handle.
func (t *Table) App() *App {
	return t.app
}

// Start runs the component.
func (t *Table) Start() {
	t.Stop()
	t.CmdBuff().AddListener(t)
	t.Styles().AddListener(t.Table)
}

// Stop terminates the component.
func (t *Table) Stop() {
	t.CmdBuff().RemoveListener(t)
	t.Styles().RemoveListener(t.Table)
}

// SetEnterFn specifies the default enter behavior.
func (t *Table) SetEnterFn(f EnterFunc) {
	t.enterFn = f
}

// SetExtraActionsFn specifies custom keyboard behavior.
func (t *Table) SetExtraActionsFn(BoostActionsFunc) {}

// BufferCompleted indicates input was accepted.
func (t *Table) BufferCompleted(text, _ string) {
	t.app.QueueUpdateDraw(func() {
		t.Filter(text)
	})
}

// BufferChanged indicates the buffer was changed.
func (t *Table) BufferChanged(_, _ string) {}

// BufferActive indicates the buff activity changed.
func (t *Table) BufferActive(state bool, k model.BufferKind) {
	t.app.BufferActive(state, k)
	if !state {
		t.app.SetFocus(t)
	}
}

func (t *Table) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveTable(t.app.Config.K9s.ContextScreenDumpDir(), t.GVR().R(), t.Path, t.GetFilteredData()); err != nil {
		t.app.Flash().Err(err)
	} else {
		t.app.Flash().Infof("File saved successfully: %q", render.Truncate(filepath.Base(path), 50))
	}

	return nil
}

func (t *Table) bindKeys() {
	t.Actions().Bulk(ui.KeyMap{
		ui.KeyHelp:             ui.NewKeyAction("Help", t.App().helpCmd, true),
		ui.KeySpace:            ui.NewSharedKeyAction("Mark", t.markCmd, false),
		tcell.KeyCtrlSpace:     ui.NewSharedKeyAction("Mark Range", t.markSpanCmd, false),
		tcell.KeyCtrlBackslash: ui.NewSharedKeyAction("Marks Clear", t.clearMarksCmd, false),
		tcell.KeyCtrlS:         ui.NewSharedKeyAction("Save", t.saveCmd, false),
		ui.KeySlash:            ui.NewSharedKeyAction("Filter Mode", t.activateCmd, false),
		tcell.KeyCtrlZ:         ui.NewKeyAction("Toggle Faults", t.toggleFaultCmd, false),
		tcell.KeyCtrlW:         ui.NewKeyAction("Toggle Wide", t.toggleWideCmd, false),
		ui.KeyShiftN:           ui.NewKeyAction("Sort Name", t.SortColCmd(nameCol, true), false),
		ui.KeyShiftA:           ui.NewKeyAction("Sort Age", t.SortColCmd(ageCol, true), false),
	})
}

func (t *Table) toggleFaultCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.ToggleToast()
	return nil
}

func (t *Table) toggleWideCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.ToggleWide()
	return nil
}

func (t *Table) cpCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := t.GetSelectedItem()
	if path == "" {
		return evt
	}
	_, n := client.Namespaced(path)
	if err := clipboardWrite(n); err != nil {
		t.app.Flash().Err(err)
		return nil
	}
	t.app.Flash().Info("Resource name copied to clipboard...")

	return nil
}

func (t *Table) cpNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := t.GetSelectedItem()
	if path == "" {
		return evt
	}
	ns, _ := client.Namespaced(path)
	if err := clipboardWrite(ns); err != nil {
		t.app.Flash().Err(err)
		return nil
	}
	t.app.Flash().Info("Resource namespace copied to clipboard...")

	return nil
}

func (t *Table) markCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.ToggleMark()
	t.Refresh()

	return nil
}

func (t *Table) markSpanCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.SpanMark()
	t.Refresh()

	return nil
}

func (t *Table) clearMarksCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.ClearMarks()
	t.Refresh()

	return nil
}

func (t *Table) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if t.app.InCmdMode() {
		return evt
	}
	t.App().ResetPrompt(t.CmdBuff())

	return evt
}
