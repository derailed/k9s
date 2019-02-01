package views

import (
	"context"
	"strings"
	"time"

	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	flashInfo flashLevel = iota
	flashWarn
	flashErr
	flashFatal
	flashDelay = 5

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
		app    *tview.Application
	}
)

func newFlashView(app *tview.Application, m string) *flashView {
	var f flashView
	{
		f = flashView{app: app, TextView: tview.NewTextView()}
		f.SetTextColor(tcell.ColorAqua)
		f.SetTextAlign(tview.AlignLeft)
		f.SetBorderPadding(0, 0, 1, 1)
	}
	return &f
}

func (f *flashView) setMessage(level flashLevel, msg ...string) {
	if f.cancel != nil {
		f.cancel()
	}

	var ctx context.Context
	{
		ctx, f.cancel = context.WithTimeout(context.TODO(), flashDelay*time.Second)
		go func(ctx context.Context) {
			m := strings.Join(msg, " ")
			f.SetTextColor(flashColor(level))
			f.SetText(flashEmoji(level) + "  " + m)
			f.app.Draw()
			for {
				select {
				case <-ctx.Done():
					f.Clear()
					f.app.Draw()
					return
				}
			}
		}(ctx)
	}
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
