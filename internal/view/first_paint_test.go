// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func firstPaintApp(t *testing.T, splashless bool) *App {
	t.Helper()

	a := NewApp(mock.NewMockConfig(t))
	a.Config.K9s.UI.Splashless = splashless
	a.Main.AddPage("main", tview.NewBox(), true, false)
	if !splashless {
		a.Main.AddPage("splash", tview.NewBox(), true, true)
	}

	return a
}

func frontPage(a *App) string {
	name, _ := a.Main.GetFrontPage()
	return name
}

func TestContentReadyShowsMainWhenSplashless(t *testing.T) {
	a := firstPaintApp(t, true)
	assert.Empty(t, frontPage(a), "nothing is drawn before the first content")

	a.contentReady()
	assert.Equal(t, "main", frontPage(a))
}

func TestContentReadyShowsMainOnce(t *testing.T) {
	a := firstPaintApp(t, true)
	a.contentReady()
	a.Main.AddPage("dialog", tview.NewBox(), true, true)

	a.contentReady()
	a.showMain()
	assert.Equal(t, "dialog", frontPage(a), "later content does not switch pages again")
}

func TestContentReadyKeepsSplash(t *testing.T) {
	a := firstPaintApp(t, false)

	a.contentReady()
	assert.Equal(t, "splash", frontPage(a), "the splash keeps its own delay")

	a.showMain()
	assert.Equal(t, "main", frontPage(a))
}
