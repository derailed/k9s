package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	// FlashInfo represents an info message.
	FlashInfo FlashLevel = iota
	// FlashWarn represents an warning message.
	FlashWarn
	// FlashErr represents an error message.
	FlashErr
	// FlashFatal represents an fatal message.
	FlashFatal

	flashDelay = 3

	emoDoh   = "😗"
	emoRed   = "😡"
	emoDead  = "💀"
	emoHappy = "😎"
)

type (
	// FlashLevel represents flash message severity.
	FlashLevel int

	// Flash represents a flash message indicator.
	Flash struct {
		*tview.TextView

		cancel context.CancelFunc
		app    *App
	}
)

// NewFlash returns a new flash view.
func NewFlash(app *App, m string) *Flash {
	f := Flash{app: app, TextView: tview.NewTextView()}
	f.SetTextColor(tcell.ColorAqua)
	f.SetTextAlign(tview.AlignLeft)
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetText("")
	f.app.Styles.AddListener(&f)

	return &f
}

// StylesChanged notifies listener the skin changed.
func (f *Flash) StylesChanged(s *config.Styles) {
	f.SetBackgroundColor(s.BgColor())
	f.SetTextColor(s.FgColor())
}

// Info displays an info flash message.
func (f *Flash) Info(msg string) {
	f.setMessage(FlashInfo, msg)
}

// Infof displays a formatted info flash message.
func (f *Flash) Infof(fmat string, args ...interface{}) {
	f.Info(fmt.Sprintf(fmat, args...))
}

// Warn displays a warning flash message.
func (f *Flash) Warn(msg string) {
	f.setMessage(FlashWarn, msg)
}

// Warnf displays a formatted warning flash message.
func (f *Flash) Warnf(fmat string, args ...interface{}) {
	f.Warn(fmt.Sprintf(fmat, args...))
}

// Err displays an error flash message.
func (f *Flash) Err(err error) {
	log.Error().Err(err).Msgf("%v", err)
	f.setMessage(FlashErr, err.Error())
}

// Errf displays a formatted error flash message.
func (f *Flash) Errf(fmat string, args ...interface{}) {
	var err error
	for _, a := range args {
		switch e := a.(type) {
		case error:
			err = e
		}
	}
	log.Error().Err(err).Msgf(fmat, args...)
	f.setMessage(FlashErr, fmt.Sprintf(fmat, args...))
}

func (f *Flash) setMessage(level FlashLevel, msg ...string) {
	if f.cancel != nil {
		f.cancel()
	}
	var ctx1, ctx2 context.Context
	{
		var timerCancel context.CancelFunc
		ctx1, f.cancel = context.WithCancel(context.TODO())
		ctx2, timerCancel = context.WithTimeout(context.TODO(), flashDelay*time.Second)
		go f.refresh(ctx1, ctx2, timerCancel)
	}
	_, _, width, _ := f.GetRect()
	if width <= 15 {
		width = 100
	}
	m := strings.Join(msg, " ")
	f.SetTextColor(flashColor(level))
	f.SetText(render.Truncate(flashEmoji(level)+" "+m, width-3))
}

func (f *Flash) refresh(ctx1, ctx2 context.Context, cancel context.CancelFunc) {
	defer cancel()
	for {
		select {
		// Timer canceled bail now
		case <-ctx1.Done():
			return
		// Timed out clear and bail
		case <-ctx2.Done():
			f.app.QueueUpdateDraw(func() {
				f.Clear()
			})
			return
		}
	}
}

func flashEmoji(l FlashLevel) string {
	switch l {
	case FlashWarn:
		return emoDoh
	case FlashErr:
		return emoRed
	case FlashFatal:
		return emoDead
	default:
		return emoHappy
	}
}

func flashColor(l FlashLevel) tcell.Color {
	switch l {
	case FlashWarn:
		return tcell.ColorOrange
	case FlashErr:
		return tcell.ColorOrangeRed
	case FlashFatal:
		return tcell.ColorFuchsia
	default:
		return tcell.ColorNavajoWhite
	}
}
