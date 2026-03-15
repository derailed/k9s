// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"log/slog"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const (
	emoHappy = "😎"
	emoDoh   = "😗"
	emoRed   = "😡"
)

// Flash represents a flash message indicator.
type Flash struct {
	*tview.TextView

	app      *App
	testMode bool
}

// NewFlash returns a new flash view.
func NewFlash(app *App) *Flash {
	f := Flash{
		app:      app,
		TextView: tview.NewTextView(),
	}
	f.SetTextColor(tcell.ColorAqua)
	f.SetDynamicColors(true)
	f.SetTextAlign(tview.AlignCenter)
	f.SetBorderPadding(0, 0, 1, 1)
	f.app.Styles.AddListener(&f)

	return &f
}

// SetTestMode for testing ONLY!
func (f *Flash) SetTestMode(b bool) {
	f.testMode = b
}

// StylesChanged notifies listener the skin changed.
func (f *Flash) StylesChanged(s *config.Styles) {
	f.SetBackgroundColor(s.BgColor())
	f.SetTextColor(s.FgColor())
}

// Watch watches for flash changes.
func (f *Flash) Watch(ctx context.Context, c model.FlashChan) {
	defer slog.Debug("Flash Watch Canceled!")
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-c:
			f.SetMessage(msg)
		}
	}
}

// SetMessage sets flash message and level.
func (f *Flash) SetMessage(m model.LevelMessage) {
	fn := func() {
		if m.Text == "" {
			f.Clear()
			return
		}
		f.SetTextColor(f.flashColor(m.Level))
		f.SetText(f.flashEmoji(m.Level) + " " + m.Text)
	}

	if f.testMode {
		fn()
	} else {
		f.app.QueueUpdateDraw(fn)
	}
}

func (f *Flash) flashEmoji(l model.FlashLevel) string {
	if f.app.Config.K9s.UI.NoIcons {
		return ""
	}
	//nolint:exhaustive
	switch l {
	case model.FlashWarn:
		return emoDoh
	case model.FlashErr:
		return emoRed
	default:
		return emoHappy
	}
}

// Helpers...

func (f *Flash) flashColor(l model.FlashLevel) tcell.Color {
	//nolint:exhaustive
	switch l {
	case model.FlashWarn:
		return f.app.Styles.Flash().WarnColor.Color()
	case model.FlashErr:
		return f.app.Styles.Flash().ErrorColor.Color()
	default:
		return f.app.Styles.Flash().InfoColor.Color()
	}
}
