// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
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

type fakeComp struct {
	*tview.Box

	name          string
	starts, stops int
	running       bool
}

func newFake(n string) *fakeComp { return &fakeComp{Box: tview.NewBox(), name: n} }

func (f *fakeComp) Name() string                         { return f.name }
func (f *fakeComp) Start()                               { f.starts++; f.running = true }
func (f *fakeComp) Stop()                                { f.stops++; f.running = false }
func (*fakeComp) Init(context.Context) error             { return nil }
func (*fakeComp) Hints() model.MenuHints                 { return nil }
func (*fakeComp) ExtraHints() map[string]string          { return nil }
func (*fakeComp) InCmdMode() bool                        { return false }
func (*fakeComp) SetCommand(*cmd.Interpreter)            {}
func (*fakeComp) SetFilter(string, bool)                 {}
func (*fakeComp) SetLabelSelector(labels.Selector, bool) {}

func newTabApp(t *testing.T) *App {
	t.Helper()

	a := NewApp(mock.NewMockConfig(t))
	ctx := context.WithValue(context.Background(), internal.KeyApp, a)
	require.NoError(t, a.Content.Init(ctx))
	a.Content.AddListener(a.Crumbs())
	a.Content.AddListener(a.Menu())
	a.Content.AddListener(a.tabBar)

	return a
}

func TestTabSwitchKeepsPerTabState(t *testing.T) {
	a := newTabApp(t)
	x1, x2, y1 := newFake("x1"), newFake("x2"), newFake("y1")
	a.Content.Push(x1)
	a.Content.Push(x2)
	a.cmdHistory.Push("pods")
	first := a.Content

	require.NoError(t, a.addTab())
	assert.Equal(t, 1, a.curTab)
	assert.NotSame(t, first, a.Content)
	assert.False(t, x2.running, "previous tab must be stopped")
	assert.Empty(t, a.cmdHistory.List())
	a.Content.Push(y1)
	a.filterHistory.Push("fred")
	assert.True(t, y1.running)
	assert.Contains(t, a.Crumbs().GetText(true), "y1")
	assert.NotContains(t, a.Crumbs().GetText(true), "x1")

	a.activateTab(0)
	assert.Same(t, first, a.Content)
	assert.Equal(t, []string{"pods"}, a.cmdHistory.List())
	assert.Empty(t, a.filterHistory.List())
	assert.True(t, x2.running)
	assert.False(t, x1.running)
	assert.False(t, y1.running)
	assert.Contains(t, a.Crumbs().GetText(true), "x2")
	assert.NotContains(t, a.Crumbs().GetText(true), "y1")
	assert.Equal(t, []string{"x1", "x2"}, a.Content.Flatten())

	// Navigation after switching still reaches every listener.
	a.Content.Pop()
	assert.True(t, x1.running)
	assert.NotContains(t, a.Crumbs().GetText(true), "x2")
}

func TestTabActivateNoop(t *testing.T) {
	a := newTabApp(t)
	x := newFake("x")
	a.Content.Push(x)

	a.activateTab(0)
	a.activateTab(-1)
	a.activateTab(5)

	assert.Equal(t, 1, x.starts)
	assert.Equal(t, 0, x.stops)
}

func TestTabLimit(t *testing.T) {
	a := newTabApp(t)
	for range maxTabs - 1 {
		require.NoError(t, a.addTab())
	}
	assert.Len(t, a.tabs, maxTabs)
	require.Error(t, a.addTab())
	assert.Len(t, a.tabs, maxTabs)
	assert.Equal(t, maxTabs-1, a.curTab)
}

func TestTabBarVisibility(t *testing.T) {
	a := newTabApp(t)
	scr := tcell.NewSimulationScreen("")
	require.NoError(t, scr.Init())
	barHeight := func() int {
		a.body.SetRect(0, 0, 80, 24)
		a.body.Draw(scr)
		_, _, _, h := a.tabBar.GetRect()
		return h
	}
	assert.Zero(t, barHeight())

	require.NoError(t, a.addTab())
	a.Content.Push(newFake("pods"))
	assert.Equal(t, 1, barHeight())
	assert.Contains(t, a.tabBar.GetText(true), "2:pods")

	require.NoError(t, a.closeTab())
	assert.Zero(t, barHeight())
}

func TestCloseActiveTab(t *testing.T) {
	a := newTabApp(t)
	x := newFake("x")
	a.Content.Push(x)
	require.NoError(t, a.addTab())
	y1, y2 := newFake("y1"), newFake("y2")
	a.Content.Push(y1)
	a.Content.Push(y2)
	closed := a.Content

	require.NoError(t, a.closeTab())

	assert.Len(t, a.tabs, 1)
	assert.Equal(t, 0, a.curTab)
	assert.True(t, closed.Empty())
	assert.False(t, y1.running)
	assert.False(t, y2.running)
	assert.True(t, x.running)
	assert.Equal(t, []string{"x"}, a.Content.Flatten())
	assert.Contains(t, a.Crumbs().GetText(true), "x")
	assert.NotContains(t, a.Crumbs().GetText(true), "y1")

	require.Error(t, a.closeTab(), "last tab must survive")
	assert.Len(t, a.tabs, 1)
}

func TestCloseTabKeepsActiveIndex(t *testing.T) {
	a := newTabApp(t)
	require.NoError(t, a.addTab())
	require.NoError(t, a.addTab())
	a.activateTab(1)
	mid := a.Content

	a.activateTab(2)
	require.NoError(t, a.closeTab()) // closes tab 2, falls back to tab 1
	assert.Equal(t, 1, a.curTab)
	assert.Same(t, mid, a.Content)

	a.activateTab(0)
	a.removeTab(1)
	assert.Equal(t, 0, a.curTab)
}

func TestRemoveInactiveTabDoesNotRestartViews(t *testing.T) {
	a := newTabApp(t)
	x1, x2 := newFake("x1"), newFake("x2")
	a.Content.Push(x1)
	a.Content.Push(x2)
	require.NoError(t, a.addTab())
	y := newFake("y")
	a.Content.Push(y)
	starts := x1.starts

	a.removeTab(0)

	assert.Equal(t, starts, x1.starts, "closing a hidden tab must not start its views")
	assert.False(t, x1.running)
	assert.False(t, x2.running)
	assert.True(t, y.running)
	assert.Equal(t, 0, a.curTab)
	assert.Len(t, a.tabs, 1)
	assert.Contains(t, a.Crumbs().GetText(true), "y")
}

func TestResetTabsKeepsActiveOnly(t *testing.T) {
	a := newTabApp(t)
	x := newFake("x")
	a.Content.Push(x)
	require.NoError(t, a.addTab())
	y := newFake("y")
	a.Content.Push(y)
	require.NoError(t, a.addTab())
	z := newFake("z")
	a.Content.Push(z)
	a.activateTab(1)
	active := a.Content

	a.resetTabs()

	assert.Len(t, a.tabs, 1)
	assert.Equal(t, 0, a.curTab)
	assert.Same(t, active, a.Content)
	assert.True(t, y.running)
	assert.False(t, x.running)
	assert.False(t, z.running)
}

func TestTabKeys(t *testing.T) {
	a := newTabApp(t)
	a.bindKeys()
	require.NoError(t, a.addTab())
	require.NoError(t, a.addTab())

	for i := range maxTabs {
		evt := tcell.NewEventKey(tcell.KeyRune, rune('1'+i), tcell.ModAlt)
		_, ok := a.HasAction(ui.AsKey(evt))
		assert.True(t, ok, "Alt-%d", i+1)
	}

	evt := tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModAlt)
	ka, _ := a.HasAction(ui.AsKey(evt))
	assert.Nil(t, ka.Action(evt))
	assert.Equal(t, 0, a.curTab)

	evt = tcell.NewEventKey(tcell.KeyRune, '9', tcell.ModAlt)
	ka, _ = a.HasAction(ui.AsKey(evt))
	assert.Nil(t, ka.Action(evt), "missing tab is swallowed")
	assert.Equal(t, 0, a.curTab)
}

func TestTabKeysIgnorePlainRunes(t *testing.T) {
	a := newTabApp(t)
	a.bindKeys()
	require.NoError(t, a.addTab())

	// 'Ä' (U+00C4) encodes to the same key code as Alt-1.
	evt := tcell.NewEventKey(tcell.KeyRune, 'Ä', tcell.ModNone)
	ka, ok := a.HasAction(ui.AsKey(evt))
	require.True(t, ok)
	assert.Same(t, evt, ka.Action(evt))
	assert.Equal(t, 1, a.curTab)
}

func TestTabKeysIgnoredInCmdMode(t *testing.T) {
	a := newTabApp(t)
	a.bindKeys()
	require.NoError(t, a.addTab())
	a.Prompt().SetModel(a.CmdBuff())
	a.CmdBuff().SetActive(true)

	evt := tcell.NewEventKey(tcell.KeyRune, '1', tcell.ModAlt)
	ka, _ := a.HasAction(ui.AsKey(evt))
	assert.Same(t, evt, ka.Action(evt))
	assert.Equal(t, 1, a.curTab)
}

func TestTabBarMouse(t *testing.T) {
	a := newTabApp(t)
	scr := tcell.NewSimulationScreen("")
	require.NoError(t, scr.Init())
	a.Content.Push(newFake("a"))
	require.NoError(t, a.addTab())
	a.Content.Push(newFake("b"))
	require.NoError(t, a.addTab())
	a.Content.Push(newFake("c"))
	a.body.SetRect(0, 0, 80, 24)
	a.body.Draw(scr)

	click := func(x, y int) bool {
		evt := tcell.NewEventMouse(x, y, tcell.Button1, 0)
		ok, _ := a.tabBar.MouseHandler()(tview.MouseLeftClick, evt, func(tview.Primitive) {})
		return ok
	}
	ix, iy, _, _ := a.tabBar.GetInnerRect()

	// Labels are " N:name " (5 cells) separated by one cell.
	assert.True(t, click(ix, iy))
	assert.Equal(t, 0, a.curTab)
	assert.True(t, click(ix+6, iy))
	assert.Equal(t, 1, a.curTab)
	assert.True(t, click(ix+12, iy))
	assert.Equal(t, 2, a.curTab)

	assert.False(t, click(ix+5, iy), "separator is not a tab")
	assert.False(t, click(ix+40, iy), "blank space is not a tab")
	assert.False(t, click(ix, iy+3), "outside the bar")
	assert.Equal(t, 2, a.curTab)
}

func TestTabBarMouseIgnoredInCmdMode(t *testing.T) {
	a := newTabApp(t)
	scr := tcell.NewSimulationScreen("")
	require.NoError(t, scr.Init())
	require.NoError(t, a.addTab())
	a.body.SetRect(0, 0, 80, 24)
	a.body.Draw(scr)
	a.Prompt().SetModel(a.CmdBuff())
	a.CmdBuff().SetActive(true)

	ix, iy, _, _ := a.tabBar.GetInnerRect()
	ok, _ := a.tabBar.MouseHandler()(tview.MouseLeftClick, tcell.NewEventMouse(ix, iy, tcell.Button1, 0), func(tview.Primitive) {})
	assert.False(t, ok)
	assert.Equal(t, 1, a.curTab)
}
