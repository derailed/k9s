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

func TestHeaderNarrowTerminalPreservesLogoAndStatus(t *testing.T) {
	a, main, screen := newHeaderTestApp(t)
	for _, art := range []string{"", "CUSTOM"} {
		a.Logo().SetLogo(art)
		a.Logo().Warn("Dial K8s Toast")
		for _, menu := range []string{"Aliases", strings.Repeat("X", 50)} {
			a.Menu().SetCell(0, 0, tview.NewTableCell(menu))
			for _, width := range []int{120, 80, 120} {
				drawHeaderTest(main, screen, width, 20)
				x, _, w, _ := a.Logo().GetRect()
				assert.Equal(t, ui.LogoWidth, w)
				assert.Equal(t, width-ui.LogoWidth, x)
				_, y, _, h := a.Logo().Status().GetRect()
				assert.Equal(t, ui.LogoArtHeight, y)
				assert.Equal(t, 1, h)
				var status strings.Builder
				for col := x; col < x+w; col++ {
					r, _, _, _ := screen.GetContent(col, y)
					status.WriteRune(r)
				}
				assert.Contains(t, status.String(), "Dial K8s Toast")
			}
		}
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
