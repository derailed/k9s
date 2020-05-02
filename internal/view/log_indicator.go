package view

import (
	"fmt"
	"sync/atomic"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// LogIndicator represents a log view indicator.
type LogIndicator struct {
	*tview.TextView

	styles       *config.Styles
	scrollStatus int32
	fullScreen   bool
	textWrap     bool
	showTime     bool
}

// NewLogIndicator returns a new indicator.
func NewLogIndicator(cfg *config.Config, styles *config.Styles) *LogIndicator {
	l := LogIndicator{
		styles:       styles,
		TextView:     tview.NewTextView(),
		scrollStatus: 1,
		fullScreen:   cfg.K9s.Logger.FullScreenLogs,
		textWrap:     cfg.K9s.Logger.TextWrap,
		showTime:     cfg.K9s.Logger.ShowTime,
	}
	l.StylesChanged(styles)
	styles.AddListener(&l)
	l.SetTextAlign(tview.AlignCenter)
	l.SetDynamicColors(true)

	return &l
}

// StylesChanged notifies listener the skin changed.
func (l *LogIndicator) StylesChanged(styles *config.Styles) {
	l.SetBackgroundColor(styles.K9s.Views.Log.Indicator.BgColor.Color())
	l.SetTextColor(styles.K9s.Views.Log.Indicator.FgColor.Color())
}

// AutoScroll reports the current scrolling status.
func (l *LogIndicator) AutoScroll() bool {
	return atomic.LoadInt32(&l.scrollStatus) == 1
}

// Timestamp reports the current timestamp mode.
func (l *LogIndicator) Timestamp() bool {
	return l.showTime
}

// TextWrap reports the current wrap mode.
func (l *LogIndicator) TextWrap() bool {
	return l.textWrap
}

// FullScreen reports the current screen mode.
func (l *LogIndicator) FullScreen() bool {
	return l.fullScreen
}

// ToggleTimestamp toggles the current timestamp mode.
func (l *LogIndicator) ToggleTimestamp() {
	l.showTime = !l.showTime
}

// ToggleFullScreen toggles the screen mode.
func (l *LogIndicator) ToggleFullScreen() {
	l.fullScreen = !l.fullScreen
	l.Refresh()
}

// ToggleTextWrap toggles the wrap mode.
func (l *LogIndicator) ToggleTextWrap() {
	l.textWrap = !l.textWrap
	l.Refresh()
}

// ToggleAutoScroll toggles the scroll mode.
func (l *LogIndicator) ToggleAutoScroll() {
	var val int32 = 1
	if l.AutoScroll() {
		val = 0
	}
	atomic.StoreInt32(&l.scrollStatus, val)
	l.Refresh()
}

// Refresh updates the view.
func (l *LogIndicator) Refresh() {
	l.Clear()
	l.update("Autoscroll", l.AutoScroll(), true)
	l.update("FullScreen", l.fullScreen, true)
	l.update("Timestamps", l.showTime, true)
	l.update("Wrap", l.textWrap, false)
}

func (l *LogIndicator) onOff(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

func (l *LogIndicator) update(title string, state bool, pad bool) {
	const spacer = "     "
	var padding string
	if pad {
		padding = spacer
	}
	fmt.Fprintf(l, "[::b]%s: %s%s", title, l.onOff(state), padding)
}
