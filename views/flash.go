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
	f := flashView{app: app, TextView: tview.NewTextView()}
	f.SetTextColor(tcell.ColorAqua)
	f.SetTextAlign(tview.AlignLeft)
	f.SetBorderPadding(0, 0, 1, 1)
	return &f
}

func (f *flashView) setMessage(level flashLevel, msg ...string) {
	if f.cancel != nil {
		f.cancel()
	}
	ctx, cancel := context.WithTimeout(context.TODO(), flashDelay*time.Second)
	f.cancel = cancel
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

func flashEmoji(l flashLevel) string {
	switch l {
	case flashWarn:
		return "😗"
	case flashErr:
		return "😡"
	case flashFatal:
		return "💀"
	default:
		return "😎"
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
