package ui

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/resource"
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

	// FlashView represents a flash message indicator.
	FlashView struct {
		*tview.TextView

		cancel context.CancelFunc
		app    *tview.Application
	}
)

// NewFlashView returns a new flash view.
func NewFlashView(app *tview.Application, m string) *FlashView {
	f := FlashView{app: app, TextView: tview.NewTextView()}
	f.SetTextColor(tcell.ColorAqua)
	f.SetTextAlign(tview.AlignLeft)
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetText("")

	return &f
}

// Info displays an info flash message.
func (v *FlashView) Info(msg string) {
	v.setMessage(FlashInfo, msg)
}

// Infof displays a formatted info flash message.
func (v *FlashView) Infof(fmat string, args ...interface{}) {
	v.Info(fmt.Sprintf(fmat, args...))
}

// Warn displays a warning flash message.
func (v *FlashView) Warn(msg string) {
	v.setMessage(FlashWarn, msg)
}

// Warnf displays a formatted warning flash message.
func (v *FlashView) Warnf(fmat string, args ...interface{}) {
	v.Warn(fmt.Sprintf(fmat, args...))
}

// Err displays an error flash message.
func (v *FlashView) Err(err error) {
	log.Error().Err(err).Msgf("%v", err)
	v.setMessage(FlashErr, err.Error())
}

// Errf displays a formatted error flash message.
func (v *FlashView) Errf(fmat string, args ...interface{}) {
	var err error
	for _, a := range args {
		switch e := a.(type) {
		case error:
			err = e
		}
	}
	log.Error().Err(err).Msgf(fmat, args...)
	v.setMessage(FlashErr, fmt.Sprintf(fmat, args...))
}

func (v *FlashView) setMessage(level FlashLevel, msg ...string) {
	if v.cancel != nil {
		v.cancel()
	}
	var ctx1, ctx2 context.Context
	{
		var timerCancel context.CancelFunc
		ctx1, v.cancel = context.WithCancel(context.TODO())
		ctx2, timerCancel = context.WithTimeout(context.TODO(), flashDelay*time.Second)
		go v.refresh(ctx1, ctx2, timerCancel)
	}
	_, _, width, _ := v.GetRect()
	if width <= 15 {
		width = 100
	}
	m := strings.Join(msg, " ")
	v.SetTextColor(flashColor(level))
	v.SetText(resource.Truncate(flashEmoji(level)+" "+m, width-3))
}

func (v *FlashView) refresh(ctx1, ctx2 context.Context, cancel context.CancelFunc) {
	defer cancel()
	for {
		select {
		// Timer canceled bail now
		case <-ctx1.Done():
			return
		// Timed out clear and bail
		case <-ctx2.Done():
			v.app.QueueUpdateDraw(func() {
				v.Clear()
				v.app.Draw()
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
