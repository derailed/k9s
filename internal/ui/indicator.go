package ui

import (
	"context"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// IndicatorView represents a status indicator.
type IndicatorView struct {
	*tview.TextView
	app       *App
	styles    *config.Styles
	permanent string

	cancel context.CancelFunc
}

// NewIndicatorView returns a new logo.
func NewIndicatorView(app *App, styles *config.Styles) *IndicatorView {
	v := IndicatorView{
		TextView: tview.NewTextView(),
		app:      app,
		styles:   styles,
	}
	v.SetTextAlign(tview.AlignCenter)
	v.SetTextColor(tcell.ColorWhite)
	v.SetBackgroundColor(styles.BgColor())
	v.SetDynamicColors(true)

	return &v
}

// SetPermanent sets permanent title to be reset to after updates
func (v *IndicatorView) SetPermanent(info string) {
	v.permanent = info
	v.SetText(info)
}

// Reset clears out the logo view and resets colors.
func (v *IndicatorView) Reset() {
	v.Clear()
	v.SetPermanent(v.permanent)
}

// Err displays a log error state.
func (v *IndicatorView) Err(msg string) {
	v.update(msg, "orangered")
}

// Warn displays a log warning state.
func (v *IndicatorView) Warn(msg string) {
	v.update(msg, "mediumvioletred")
}

// Info displays a log info state.
func (v *IndicatorView) Info(msg string) {
	v.update(msg, "lawngreen")
}

func (v *IndicatorView) update(msg, c string) {
	v.setText(fmt.Sprintf("[%s::b] <%s> ", c, msg))
}

func (v *IndicatorView) setText(msg string) {
	if v.cancel != nil {
		v.cancel()
	}
	v.SetText(msg)

	var ctx context.Context
	ctx, v.cancel = context.WithCancel(context.Background())
	go func(ctx context.Context) {
		select {
		case <-ctx.Done():
			return
		case <-time.After(5 * time.Second):
			v.app.QueueUpdateDraw(func() {
				v.Reset()
			})
		}
	}(ctx)
}
