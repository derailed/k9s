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
		{clusterInfoWidth + len(menuText), 0},
		{threshold + 10, ui.LogoWidth},
	} {
		drawHeaderTest(main, screen, tc.width, 20)
		_, _, width, _ := a.Logo().GetRect()
		assert.Equal(t, tc.logoWidth, width, "terminal width %d, menu width %d", tc.width, a.menuWidth())
		_, _, width, _ = a.Menu().GetRect()
		assert.GreaterOrEqual(t, width, len(menuText))
		_, _, width, _ = a.clusterInfo().GetRect()
		assert.Equal(t, clusterInfoWidth, width)
		_, _, _, artHeight := a.Logo().Logo().GetRect()
		_, statusY, _, statusHeight := a.Logo().Status().GetRect()
		assert.Equal(t, ui.LogoArtHeight, artHeight)
		assert.Equal(t, ui.LogoArtHeight, statusY)
		assert.Equal(t, 1, statusHeight)
		assert.True(t, a.showLogo)
		assert.Same(t, header, main.ItemAt(0))
	}
}

func TestHeaderStockAndCustomLayoutParity(t *testing.T) {
	for name, tc := range map[string]struct {
		width, height        int
		prompt, header, logo bool
	}{
		"normal":       {120, 20, false, true, true},
		"narrow":       {70, 20, false, true, true},
		"short":        {120, 6, false, true, true},
		"prompt":       {120, 20, true, true, true},
		"short prompt": {120, 6, true, true, true},
		"logoless":     {120, 20, true, true, false},
		"headless":     {120, 20, true, false, false},
	} {
		t.Run(name, func(t *testing.T) {
			a, main, screen := newHeaderTestApp(t)
			assert.Nil(t, main.GetDrawFunc())
			a.toggleHeader(tc.header, tc.logo)
			if tc.prompt {
				main.AddItemAtIndex(1, a.Prompt(), 3, 1, false)
			}
			geometry := func() [][4]int {
				var rects [][4]int
				for i := 0; main.ItemAt(i) != nil; i++ {
					x, y, w, h := main.ItemAt(i).GetRect()
					rects = append(rects, [4]int{x, y, w, h})
				}
				if tc.header {
					for _, p := range []tview.Primitive{a.clusterInfo(), a.Menu(), a.Logo()} {
						x, y, w, h := p.GetRect()
						rects = append(rects, [4]int{x, y, w, h})
					}
				}
				return rects
			}
			drawHeaderTest(main, screen, tc.width, tc.height)
			stock := geometry()
			header := main.ItemAt(0)
			a.Logo().SetLogo(strings.Repeat(strings.Repeat("X", 100)+"\n", 100))
			drawHeaderTest(main, screen, tc.width, tc.height)
			assert.Equal(t, stock, geometry())
			assert.Same(t, header, main.ItemAt(0))
			a.Logo().SetLogo("")
			drawHeaderTest(main, screen, tc.width, tc.height)
			assert.Equal(t, stock, geometry())
		})
	}
}

func TestHeaderStylesPreserveLayoutAndVisibility(t *testing.T) {
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
	assert.Same(t, header, main.ItemAt(0))
	assert.Contains(t, a.Logo().Logo().GetText(false), "CONTEXT")
	a.toggleHeader(true, false)
	drawHeaderTest(main, screen, 120, 20)
	header = main.ItemAt(0)
	a.ReloadStyles()
	assert.Same(t, header, main.ItemAt(0))
	assert.False(t, a.showLogo)
	assert.Nil(t, header.(*tview.Flex).ItemAt(2))
	a.toggleHeader(false, false)
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

func TestHeaderMenuChangesToggleLogo(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	header := main.ItemAt(0)
	for _, tc := range []struct {
		text  string
		width int
	}{
		{"Aliases", ui.LogoWidth},
		{strings.Repeat("X", 50), 0},
		{"Aliases", ui.LogoWidth},
		{"Aliases", ui.LogoWidth},
	} {
		a.Menu().SetCell(0, 0, tview.NewTableCell(tc.text))
		drawHeaderTest(main, screen, 120, 20)
		_, _, width, _ := a.Logo().GetRect()
		assert.Equal(t, tc.width, width)
		assert.Same(t, header, main.ItemAt(0))
	}
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
