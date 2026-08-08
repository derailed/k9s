// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/labels"
)

// ---------------------------------------------------------------------------
// Test fix #1 — outgoing top is Stop()'d on tab switch (no goroutine leak).
// ---------------------------------------------------------------------------

func TestActivateSession_StopsOutgoingTop(t *testing.T) {
	app := view_newAppForTest(t)
	require.NotNil(t, app.tabManager)
	require.Equal(t, 1, app.tabManager.Count())

	// Open a second tab so we have something to switch between.
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)
	require.NoError(t, app.tabManager.newTab(ctx))
	require.Equal(t, 2, app.tabManager.Count())

	// Push a counting component onto each tab's PageStack.  Pushing notifies
	// the PageStack listener which calls Start() — that's expected setup
	// noise; we reset counters before the actual assertions.
	c0 := newCountingComp("tab-0-top")
	c1 := newCountingComp("tab-1-top")
	app.tabManager.sessions[0].Content.Push(c0)
	app.tabManager.sessions[1].Content.Push(c1)

	// After both pushes, the active tab is sessions[1] (set by newTab above).
	// Force a known starting point and reset the counters.
	app.tabManager.SwitchTo(1)
	c0.reset()
	c1.reset()

	// Switching to tab 0 must Stop() c1 (outgoing) and Start() c0 (incoming).
	app.tabManager.SwitchTo(0)
	assert.Equal(t, int32(1), c1.stops.Load(), "outgoing top should be Stop()'d on switch")
	assert.Equal(t, int32(1), c0.starts.Load(), "incoming top should be Start()'d on switch")

	// Switching back must Stop() c0 and Start() c1.
	app.tabManager.SwitchTo(1)
	assert.Equal(t, int32(1), c0.stops.Load(), "outgoing top should be Stop()'d on switch back")
	assert.Equal(t, int32(1), c1.starts.Load(), "incoming top should be Start()'d on switch back")

	// Round-trip again — counters increment by exactly one per side.
	app.tabManager.SwitchTo(0)
	assert.Equal(t, int32(2), c1.stops.Load())
	assert.Equal(t, int32(2), c0.starts.Load())
}

// ---------------------------------------------------------------------------
// Test fix #2 — switchNS does not panic when ResourceViewer returns nil from
// GetTable() or from GetTable().GetModel().
// ---------------------------------------------------------------------------

func TestSwitchNS_NoPanicOnNilTableOrModel(t *testing.T) {
	app := view_newAppForTest(t)

	// Stub viewer with nil Table.
	nilTableViewer := &stubResourceViewer{
		name:  "nil-table",
		gvr:   client.PodGVR,
		table: nil,
	}

	// Stub viewer with non-nil Table but uninitialised (nil) embedded model.
	bareTable := &Table{Table: ui.NewTable(client.PodGVR)}
	nilModelViewer := &stubResourceViewer{
		name:  "nil-model",
		gvr:   client.PodGVR,
		table: bareTable,
	}

	// Push both onto the initial tab so switchNS will iterate them.
	app.tabManager.sessions[0].Content.Push(nilTableViewer)
	app.tabManager.sessions[0].Content.Push(nilModelViewer)

	// Must not panic.
	assert.NotPanics(t, func() {
		app.tabManager.switchNS("new-ns")
	})
}

// ---------------------------------------------------------------------------
// Test fix #4 — state consistency after closeActive (no orphan pages or
// stale active index).  Also exercises the post-mutex-removal code paths.
// ---------------------------------------------------------------------------

func TestCloseActive_StateConsistencyAfterRemovingMiddleTab(t *testing.T) {
	app := view_newAppForTest(t)
	ctx := context.WithValue(context.Background(), internal.KeyApp, app)

	// Open four extra tabs so we have five total: indices 0..4.
	for i := 0; i < 4; i++ {
		require.NoError(t, app.tabManager.newTab(ctx))
	}
	require.Equal(t, 5, app.tabManager.Count())

	// Make the middle tab active and close it.
	app.tabManager.SwitchTo(2)
	require.NoError(t, app.tabManager.closeActive())

	// Down to four tabs; the right-neighbour preference means the new active
	// index is still 2 (the tab that was previously at index 3).
	assert.Equal(t, 4, app.tabManager.Count())
	assert.Equal(t, 2, app.tabManager.activeIdx)

	// Close the rightmost tab — fallback to the left neighbour.
	app.tabManager.SwitchTo(3)
	require.NoError(t, app.tabManager.closeActive())
	assert.Equal(t, 3, app.tabManager.Count())
	assert.Equal(t, 2, app.tabManager.activeIdx)

	// Reduce to a single tab and verify last-tab close is rejected.
	require.NoError(t, app.tabManager.closeActive())
	require.NoError(t, app.tabManager.closeActive())
	assert.Equal(t, 1, app.tabManager.Count())
	err := app.tabManager.closeActive()
	assert.Error(t, err)
	assert.Equal(t, 1, app.tabManager.Count())
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// view_newAppForTest constructs an App with mock config, initialised enough
// to host a TabManager.  Mirrors the pattern used by app_test.go.
func view_newAppForTest(t *testing.T) *App {
	t.Helper()
	a := NewApp(mock.NewMockConfig(t))
	// Init may return an error in mock setup (no real cluster); ignore it
	// just like the existing TestAppNew does — what we need is tabManager
	// being constructed, which happens unconditionally near the end of Init.
	_ = a.Init("test", 0)
	require.NotNil(t, a.tabManager, "App.Init did not wire tabManager")
	require.GreaterOrEqual(t, a.tabManager.Count(), 1)
	return a
}

// countingComp is a minimal model.Component stub that counts Start/Stop calls.
// It's modelled after the `c` stub in internal/model/stack_test.go.
type countingComp struct {
	name   string
	starts atomic.Int32
	stops  atomic.Int32
}

func newCountingComp(name string) *countingComp {
	return &countingComp{name: name}
}

func (c *countingComp) reset() {
	c.starts.Store(0)
	c.stops.Store(0)
}

func (c *countingComp) Name() string                  { return c.name }
func (c *countingComp) Init(context.Context) error    { return nil }
func (c *countingComp) Start()                        { c.starts.Add(1) }
func (c *countingComp) Stop()                         { c.stops.Add(1) }
func (c *countingComp) Hints() model.MenuHints        { return nil }
func (c *countingComp) ExtraHints() map[string]string { return nil }
func (c *countingComp) InCmdMode() bool               { return false }
func (c *countingComp) SetFilter(string, bool)        {}
func (*countingComp) SetLabelSelector(labels.Selector, bool) {}
func (*countingComp) SetCommand(*cmd.Interpreter)            {}

// tview.Primitive surface.
func (*countingComp) Draw(tcell.Screen)                                              {}
func (*countingComp) GetRect() (int, int, int, int)                                  { return 0, 0, 0, 0 }
func (*countingComp) SetRect(int, int, int, int)                                     {}
func (*countingComp) InputHandler() func(*tcell.EventKey, func(tview.Primitive))     { return nil }
func (*countingComp) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return nil
}
func (*countingComp) Focus(func(tview.Primitive))   {}
func (*countingComp) Blur()                         {}
func (*countingComp) HasFocus() bool                { return false }
func (*countingComp) GetFocusable() tview.Focusable { return nil }

// stubResourceViewer implements just enough of view.ResourceViewer for
// switchNS to iterate over it and exercise the nil-Table / nil-Model paths.
type stubResourceViewer struct {
	name  string
	gvr   *client.GVR
	table *Table
}

func (s *stubResourceViewer) Name() string                  { return s.name }
func (s *stubResourceViewer) Init(context.Context) error    { return nil }
func (s *stubResourceViewer) Start()                        {}
func (s *stubResourceViewer) Stop()                         {}
func (s *stubResourceViewer) Hints() model.MenuHints        { return nil }
func (s *stubResourceViewer) ExtraHints() map[string]string { return nil }
func (s *stubResourceViewer) InCmdMode() bool               { return false }
func (s *stubResourceViewer) SetFilter(string, bool)        {}
func (*stubResourceViewer) SetLabelSelector(labels.Selector, bool) {}
func (*stubResourceViewer) SetCommand(*cmd.Interpreter)            {}

// tview.Primitive surface.
func (*stubResourceViewer) Draw(tcell.Screen)                                              {}
func (*stubResourceViewer) GetRect() (int, int, int, int)                                  { return 0, 0, 0, 0 }
func (*stubResourceViewer) SetRect(int, int, int, int)                                     {}
func (*stubResourceViewer) InputHandler() func(*tcell.EventKey, func(tview.Primitive))     { return nil }
func (*stubResourceViewer) MouseHandler() func(tview.MouseAction, *tcell.EventMouse, func(tview.Primitive)) (bool, tview.Primitive) {
	return nil
}
func (*stubResourceViewer) Focus(func(tview.Primitive))   {}
func (*stubResourceViewer) Blur()                         {}
func (*stubResourceViewer) HasFocus() bool                { return false }
func (*stubResourceViewer) GetFocusable() tview.Focusable { return nil }

// view.Viewer / TableViewer / ResourceViewer.
func (*stubResourceViewer) Actions() *ui.KeyActions { return nil }
func (*stubResourceViewer) App() *App               { return nil }
func (*stubResourceViewer) Refresh()                {}
func (s *stubResourceViewer) GetTable() *Table      { return s.table }
func (s *stubResourceViewer) GVR() *client.GVR      { return s.gvr }
func (*stubResourceViewer) SetEnvFn(EnvFunc)        {}
func (*stubResourceViewer) SetContextFn(ContextFunc) {}
func (*stubResourceViewer) AddBindKeysFn(BindKeysFunc) {}
func (*stubResourceViewer) SetInstance(string)        {}
