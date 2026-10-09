// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newHeaderTestApp(t *testing.T) (*App, *tview.Flex, tcell.SimulationScreen) {
	t.Helper()
	a := NewApp(mock.NewMockConfig(t))
	a.Config.K9s.UI.Splashless = true
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	a.layout(ctx)
	main := a.Main.GetPrimitive("main").(*tview.Flex)
	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init())
	t.Cleanup(screen.Fini)
	return a, main, screen
}

func drawHeaderTest(main *tview.Flex, screen tcell.SimulationScreen, width, height int) {
	screen.SetSize(width, height)
	screen.Clear()
	main.SetRect(0, 0, width, height)
	main.Draw(screen)
}

func TestHeaderNarrowTerminalHidesAndRestoresLogo(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	const menuText = " <ctrl-a> Aliases "
	a.Menu().SetCell(0, 0, tview.NewTableCell("[red::b]"+menuText+"[-::-]"))
	threshold := clusterInfoWidth + len(menuText) + ui.LogoWidth
	header := main.ItemAt(0)
	for _, tc := range []struct {
		width, logoWidth int
	}{
		{threshold - 1, 0},
		{threshold, ui.LogoWidth},
		{10, 0},
		{threshold + 10, ui.LogoWidth},
	} {
		drawHeaderTest(main, screen, tc.width, 20)
		_, _, width, _ := a.Logo().GetRect()
		assert.Equal(t, tc.logoWidth, width, "terminal width %d, menu width %d", tc.width, a.menuWidth())
		_, _, width, _ = a.Menu().GetRect()
		assert.GreaterOrEqual(t, width, min(tc.width, len(menuText)))
		_, _, _, artHeight := a.Logo().Logo().GetRect()
		_, statusY, _, statusHeight := a.Logo().Status().GetRect()
		assert.Equal(t, ui.LogoArtHeight, artHeight)
		assert.Equal(t, ui.LogoArtHeight, statusY)
		assert.Equal(t, 1, statusHeight)
		assert.True(t, a.showLogo)
		assert.Same(t, header, main.ItemAt(0))
	}
}

func TestHeaderOversizedArtAndShortTerminal(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	a.Logo().SetLogo(strings.Repeat(strings.Repeat("X", 100)+"\n", 100))
	for _, height := range []int{30, 10, 6, 3, 2, 1, 0, 30} {
		drawHeaderTest(main, screen, 120, height)
		_, _, _, headerHeight := main.ItemAt(0).GetRect()
		assert.LessOrEqual(t, headerHeight, ui.LogoArtHeight+1)
		_, _, _, contentHeight := a.Content.GetRect()
		assert.GreaterOrEqual(t, contentHeight, min(1, height))
		for i := 0; main.ItemAt(i) != nil; i++ {
			_, y, _, h := main.ItemAt(i).GetRect()
			assert.GreaterOrEqual(t, h, 0)
			assert.LessOrEqual(t, y+h, height)
		}
	}
}

func TestHeaderActivePromptKeepsContentVisible(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	main.AddItemAtIndex(1, a.Prompt(), 3, 1, false)
	for _, height := range []int{20, 6, 3, 1, 0} {
		drawHeaderTest(main, screen, 120, height)
		_, _, _, contentHeight := a.Content.GetRect()
		assert.GreaterOrEqual(t, contentHeight, min(1, height))
		for i := 0; main.ItemAt(i) != nil; i++ {
			_, y, _, h := main.ItemAt(i).GetRect()
			assert.GreaterOrEqual(t, h, 0)
			assert.LessOrEqual(t, y+h, height)
		}
	}
}

func TestRefreshHeaderPreservesLayoutAndVisibility(t *testing.T) {
	a := NewApp(mock.NewMockConfig(t))
	assert.NotPanics(t, a.RefreshHeader)
	a, main, screen := newHeaderTestApp(t)
	drawHeaderTest(main, screen, 120, 20)
	header := main.ItemAt(0)
	a.Config.K9s.UI.Logo = "GLOBAL"
	a.ReloadStyles()
	assert.Same(t, header, main.ItemAt(0))
	assert.Contains(t, a.Logo().Logo().GetText(false), "GLOBAL")
	ct, err := a.Config.K9s.ActivateContext("ct-1-1")
	require.NoError(t, err)
	ct.Logo = "CONTEXT"
	a.ReloadStyles()
	a.ReloadStyles()
	a.RefreshHeader()
	assert.Same(t, header, main.ItemAt(0))
	assert.Contains(t, a.Logo().Logo().GetText(false), "CONTEXT")
	a.toggleHeader(true, false)
	drawHeaderTest(main, screen, 120, 20)
	header = main.ItemAt(0)
	a.ReloadStyles()
	a.RefreshHeader()
	assert.Same(t, header, main.ItemAt(0))
	assert.False(t, a.showLogo)
	assert.Nil(t, header.(*tview.Flex).ItemAt(2))
	a.toggleHeader(false, false)
	a.RefreshHeader()
	assert.Same(t, a.statusIndicator(), main.ItemAt(0))
}

func TestHeaderMenuWidthSkipsEmptyColumns(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	a.Menu().SetCell(0, 0, tview.NewTableCell(""))
	a.Menu().SetCell(0, 1, tview.NewTableCell("[#ff0000::b]Aliases"))
	assert.Equal(t, len("Aliases"), a.menuWidth())
	drawHeaderTest(main, screen, clusterInfoWidth+len("Aliases")+ui.LogoWidth-1, 20)
	_, _, width, _ := a.Logo().GetRect()
	assert.Zero(t, width)
}

func TestHeaderHeight(t *testing.T) {
	a := &App{showHeader: true}
	for _, rows := range []int{-1, 0, 1, 6, 7, 100} {
		assert.Equal(t, min(7, max(0, rows)), a.headerHeight(rows))
	}
	a.showHeader = false
	assert.Equal(t, 1, a.headerHeight(100))
	assert.Zero(t, a.headerHeight(0))
}

func TestLogoWidth(t *testing.T) {
	a := &App{showLogo: true}
	assert.Zero(t, a.logoWidth(75, 50, 1))
	assert.Zero(t, a.logoWidth(76, 50, 1))
	assert.Equal(t, ui.LogoWidth, a.logoWidth(77, 50, 1))
	assert.Zero(t, a.logoWidth(100, 80, 1))
	assert.Zero(t, a.logoWidth(0, 50, 1))
	a.showLogo = false
	assert.Zero(t, a.logoWidth(120, 50, 1))
}
