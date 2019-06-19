package views

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
	flashInfo flashLevel = iota
	flashWarn
	flashErr
	flashFatal
	flashDelay = 3

	emoDoh   = "😗"
	emoRed   = "😡"
	emoDead  = "💀"
	emoHappy = "😎"
)

type (
	flashLevel int

	flashView struct {
		*tview.TextView

		cancel context.CancelFunc
		app    *appView
	}
)

func newFlashView(app *appView, m string) *flashView {
	f := flashView{app: app, TextView: tview.NewTextView()}
	f.SetTextColor(tcell.ColorAqua)
	f.SetTextAlign(tview.AlignLeft)
	f.SetBorderPadding(0, 0, 1, 1)
	f.SetText("")

	return &f
}

func (v *flashView) info(msg string) {
	v.setMessage(flashInfo, msg)
}

func (v *flashView) infof(fmat string, args ...interface{}) {
	v.info(fmt.Sprintf(fmat, args...))
}

func (v *flashView) warn(msg string) {
	v.setMessage(flashWarn, msg)
}

func (v *flashView) warnf(fmat string, args ...interface{}) {
	v.warn(fmt.Sprintf(fmat, args...))
}

func (v *flashView) err(err error) {
	log.Error().Err(err).Msgf("%v", err)
	v.setMessage(flashErr, err.Error())
}

func (v *flashView) errf(fmat string, args ...interface{}) {
	var err error
	for _, a := range args {
		switch e := a.(type) {
		case error:
			err = e
		}
	}
	log.Error().Err(err).Msgf(fmat, args...)
	v.setMessage(flashErr, fmt.Sprintf(fmat, args...))
}

func (v *flashView) setMessage(level flashLevel, msg ...string) {
	if v.cancel != nil {
		v.cancel()
	}
	var ctx1, ctx2 context.Context
	{
		var timerCancel context.CancelFunc
		ctx1, v.cancel = context.WithCancel(context.TODO())
		ctx2, timerCancel = context.WithTimeout(context.TODO(), flashDelay*time.Second)
		go func(ctx1, ctx2 context.Context, timerCancel context.CancelFunc) {
			defer timerCancel()
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
		}(ctx1, ctx2, timerCancel)
	}
	_, _, width, _ := v.GetRect()
	if width <= 15 {
		width = 100
	}
	m := strings.Join(msg, " ")
	v.SetTextColor(flashColor(level))
	v.SetText(resource.Truncate(flashEmoji(level)+" "+m, width-3))
}

func flashEmoji(l flashLevel) string {
	switch l {
	case flashWarn:
		return emoDoh
	case flashErr:
		return emoRed
	case flashFatal:
		return emoDead
	default:
		return emoHappy
	}
}

func flashColor(l flashLevel) tcell.Color {
	switch l {
	case flashWarn:
		return tcell.ColorOrange
	case flashErr:
		return tcell.ColorOrangeRed
	case flashFatal:
		return tcell.ColorFuchsia
	default:
		return tcell.ColorNavajoWhite
	}
}
