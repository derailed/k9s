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
}

// NewLogIndicator returns a new indicator.
func NewLogIndicator(styles *config.Styles) *LogIndicator {
	l := LogIndicator{
		styles:       styles,
		TextView:     tview.NewTextView(),
		scrollStatus: 1,
	}
	l.SetBackgroundColor(config.AsColor(styles.Views().Log.BgColor))
	l.SetTextAlign(tview.AlignRight)
	l.SetDynamicColors(true)

	return &l
}

func (l *LogIndicator) AutoScroll() bool {
	return atomic.LoadInt32(&l.scrollStatus) == 1
}

func (l *LogIndicator) TextWrap() bool {
	return l.textWrap
}

func (l *LogIndicator) FullScreen() bool {
	return l.fullScreen
}

func (l *LogIndicator) ToggleFullScreen() {
	l.fullScreen = !l.fullScreen
	l.Refresh()
}

func (l *LogIndicator) ToggleTextWrap() {
	l.textWrap = !l.textWrap
	l.Refresh()
}

func (l *LogIndicator) ToggleAutoScroll() {
	var val int32 = 1
	if l.AutoScroll() {
		val = 0
	}
	atomic.StoreInt32(&l.scrollStatus, val)
	l.Refresh()
}

func (l *LogIndicator) Refresh() {
	l.Clear()
	l.update("Autoscroll: " + l.onOff(l.AutoScroll()))
	l.update("FullScreen: " + l.onOff(l.fullScreen))
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
