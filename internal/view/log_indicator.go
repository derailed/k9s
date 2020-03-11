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
		fullScreen:   cfg.K9s.FullScreenLogs,
	}
	l.SetBackgroundColor(styles.Views().Log.BgColor.Color())
	l.SetTextAlign(tview.AlignRight)
	l.SetDynamicColors(true)

	return &l
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
	l.update("Autoscroll: " + l.onOff(l.AutoScroll()))
	l.update("FullScreen: " + l.onOff(l.fullScreen))
	// BOZO!! log timestamp
	// l.update("Timestamp: " + l.onOff(l.showTime))
	l.update("Wrap: " + l.onOff(l.textWrap))
}

func (l *LogIndicator) onOff(b bool) string {
	if b {
		return "On"
	}
	return "Off"
}

func (l *LogIndicator) update(status string) {
	fg, bg := l.styles.Frame().Crumb.FgColor, l.styles.Frame().Crumb.ActiveColor
	fmt.Fprintf(l, "[%s:%s:b] %-15s ", fg, bg, status)
}
