// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"strings"
	"sync"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogoView(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	v.Reset()

	const elogo = "[#ffa500::b] ____  __ ________       \n[#ffa500::b]|    |/  /   __   \\______\n[#ffa500::b]|       /\\____    /  ___/\n[#ffa500::b]|    \\   \\  /    /\\___  \\\n[#ffa500::b]|____|\\__ \\/____//____  /\n[#ffa500::b]         \\/           \\/ \n"
	assert.Equal(t, elogo, v.Logo().GetText(false))
	assert.Empty(t, v.Status().GetText(false))
}

func TestLogoCustomArt(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	v.SetLogo("AB\n C")

	assert.Equal(t, "[#ffa500::b]AB\n[#ffa500::b] C\n", v.Logo().GetText(false))
}

func TestLogoRendering(t *testing.T) {
	for name, tc := range map[string]struct {
		art  string
		rows []string
	}{
		"brackets":             {"[red] [x] [::b]", []string{"[red] [x] [::b]"}},
		"tabs and indentation": {" A\tB", []string{" A    B"}},
		"CRLF":                 {" A\r\nB\r\n", []string{" A", "B"}},
		"wide art":             {strings.Repeat("A", 30), []string{strings.Repeat("A", ui.LogoWidth)}},
		"tall art":             {"1\n2\n3\n4\n5\n6\n7\n8", []string{"1", "2", "3", "4", "5", "6"}},
	} {
		t.Run(name, func(t *testing.T) {
			screen := tcell.NewSimulationScreen("UTF-8")
			require.NoError(t, screen.Init())
			defer screen.Fini()
			screen.SetSize(ui.LogoWidth, ui.LogoArtHeight+1)
			v := ui.NewLogo(config.NewStyles())
			v.SetRect(0, 0, ui.LogoWidth, ui.LogoArtHeight+1)
			v.SetLogo(tc.art)
			v.Draw(screen)
			for y, want := range tc.rows {
				var row strings.Builder
				for x := range ui.LogoWidth {
					r, _, style, _ := screen.GetContent(x, y)
					row.WriteRune(r)
					if x < len(want) {
						fg, _, _ := style.Decompose()
						assert.Equal(t, config.NewStyles().Body().LogoColor.Color(), fg)
					}
				}
				assert.Equal(t, want, strings.TrimRight(row.String(), " "))
			}
		})
	}
}

func TestLogoWideRunesAndCombiningCharacters(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	screen.SetSize(ui.LogoWidth, ui.LogoArtHeight+1)
	v := ui.NewLogo(config.NewStyles())
	v.SetRect(0, 0, ui.LogoWidth, ui.LogoArtHeight+1)
	v.SetLogo(strings.Repeat("\u754c", 13) + "X\ne\u0301\t\u754c")
	v.Draw(screen)
	for x := 0; x < ui.LogoWidth; x += 2 {
		r, _, _, width := screen.GetContent(x, 0)
		assert.Equal(t, '\u754c', r)
		assert.Equal(t, 2, width)
	}
	r, combining, _, _ := screen.GetContent(0, 1)
	assert.Equal(t, 'e', r)
	assert.Equal(t, []rune{'\u0301'}, combining)
	r, _, _, _ = screen.GetContent(5, 1)
	assert.Equal(t, '\u754c', r)
}

func TestLogoShortPanel(t *testing.T) {
	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	screen.SetSize(ui.LogoWidth, 3)
	v := ui.NewLogo(config.NewStyles())
	v.SetLogo("A\nB\nC\nD")
	v.Info("status")
	v.SetRect(0, 0, ui.LogoWidth, 3)
	v.Draw(screen)
	_, _, _, artHeight := v.Logo().GetRect()
	_, statusY, _, statusHeight := v.Status().GetRect()
	assert.Equal(t, 2, artHeight)
	assert.Equal(t, 2, statusY)
	assert.Equal(t, 1, statusHeight)
	r, _, _, _ := screen.GetContent(0, 0)
	assert.Equal(t, 'A', r)
}

func TestLogoConcurrentUpdates(t *testing.T) {
	styles := config.NewStyles()
	v := ui.NewLogo(styles)
	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	screen.SetSize(ui.LogoWidth, ui.LogoArtHeight+1)
	v.SetRect(0, 0, ui.LogoWidth, ui.LogoArtHeight+1)
	var wg sync.WaitGroup
	for _, update := range []func(){
		func() { v.SetLogo("[red]\t\u754c\nsecond") },
		func() { v.StylesChanged(styles) },
		func() { v.Info("Bench"); v.Warn("warn"); v.Err("err"); v.Reset() },
		func() { v.IsBenchmarking() },
		func() { v.Draw(screen) },
	} {
		wg.Go(func() {
			for range 100 {
				update()
			}
		})
	}
	wg.Wait()
}

func TestLogoCustomArtFallsBack(t *testing.T) {
	v := ui.NewLogo(config.NewStyles())
	v.SetLogo(" \n\t\n")

	const elogo = "[#ffa500::b] ____  __ ________       \n[#ffa500::b]|    |/  /   __   \\______\n[#ffa500::b]|       /\\____    /  ___/\n[#ffa500::b]|    \\   \\  /    /\\___  \\\n[#ffa500::b]|____|\\__ \\/____//____  /\n[#ffa500::b]         \\/           \\/ \n"
	assert.Equal(t, elogo, v.Logo().GetText(false))
}

func TestLogoStatus(t *testing.T) {
	uu := map[string]struct {
		logo, msg, e string
	}{
		"info": {
			"[#008000::b] ____  __ ________       \n[#008000::b]|    |/  /   __   \\______\n[#008000::b]|       /\\____    /  ___/\n[#008000::b]|    \\   \\  /    /\\___  \\\n[#008000::b]|____|\\__ \\/____//____  /\n[#008000::b]         \\/           \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"warn": {
			"[#c71585::b] ____  __ ________       \n[#c71585::b]|    |/  /   __   \\______\n[#c71585::b]|       /\\____    /  ___/\n[#c71585::b]|    \\   \\  /    /\\___  \\\n[#c71585::b]|____|\\__ \\/____//____  /\n[#c71585::b]         \\/           \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
		"err": {
			"[#ff0000::b] ____  __ ________       \n[#ff0000::b]|    |/  /   __   \\______\n[#ff0000::b]|       /\\____    /  ___/\n[#ff0000::b]|    \\   \\  /    /\\___  \\\n[#ff0000::b]|____|\\__ \\/____//____  /\n[#ff0000::b]         \\/           \\/ \n",
			"blee",
			"[#ffffff::b]blee\n",
		},
	}

	v := ui.NewLogo(config.NewStyles())
	for n := range uu {
		k, u := n, uu[n]
		t.Run(k, func(t *testing.T) {
			switch k {
			case "info":
				v.Info(u.msg)
			case "warn":
				v.Warn(u.msg)
			case "err":
				v.Err(u.msg)
			}
			assert.Equal(t, u.logo, v.Logo().GetText(false))
			assert.Equal(t, u.e, v.Status().GetText(false))
		})
	}
}
