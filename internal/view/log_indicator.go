package view

import (
	"sync/atomic"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

const (
	autoscroll    = "Autoscroll"
	fullscreen    = "FullScreen"
	timestamp     = "Timestamps"
	wrap          = "Wrap"
	allContainers = "AllContainers"
	on            = "On"
	off           = "Off"
	spacer        = "     "
	bold          = "[::b]"
)

// LogIndicator represents a log view indicator.
type LogIndicator struct {
	*tview.TextView

	styles                     *config.Styles
	scrollStatus               int32
	fullScreen                 bool
	textWrap                   bool
	showTime                   bool
	allContainers              bool
	shouldDisplayAllContainers bool
}

// NewLogIndicator returns a new indicator.
func NewLogIndicator(cfg *config.Config, styles *config.Styles, isContainerLogView bool) *LogIndicator {
	l := LogIndicator{
		styles:                     styles,
		TextView:                   tview.NewTextView(),
		scrollStatus:               1,
		fullScreen:                 cfg.K9s.Logger.FullScreenLogs,
		textWrap:                   cfg.K9s.Logger.TextWrap,
		showTime:                   cfg.K9s.Logger.ShowTime,
		shouldDisplayAllContainers: isContainerLogView,
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

// ToggleTextWrap toggles the wrap mode.
func (l *LogIndicator) ToggleAllContainers() {
	l.allContainers = !l.allContainers
	l.Refresh()
}

// Refresh updates the view.
func (l *LogIndicator) Refresh() {
	l.Clear()
	if l.shouldDisplayAllContainers {
		l.update(allContainers, l.allContainers, spacer)
	}
	l.update(autoscroll, l.AutoScroll(), spacer)
	l.update(fullscreen, l.fullScreen, spacer)
	l.update(timestamp, l.showTime, spacer)
	l.update(wrap, l.textWrap, "")
}

func (l *LogIndicator) update(title string, state bool, padding string) {
	bb := []byte(bold + title + ":")
	if state {
		bb = append(bb, []byte(on)...)
	} else {
		bb = append(bb, []byte(off)...)
	}
	_, _ = l.Write(append(bb, []byte(padding)...))
}
