// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"
	"sync/atomic"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

const spacer = "     "

// LogIndicator represents a log view indicator.
type LogIndicator struct {
	*tview.TextView

	styles                     *config.Styles
	scrollStatus               int32
	indicator                  []byte
	fullScreen                 bool
	textWrap                   bool
	showTime                   bool
	allContainers              bool
	shouldDisplayAllContainers bool
}

// NewLogIndicator returns a new indicator.
func NewLogIndicator(cfg *config.Config, styles *config.Styles, allContainers bool) *LogIndicator {
	l := LogIndicator{
		styles:                     styles,
		TextView:                   tview.NewTextView(),
		indicator:                  make([]byte, 0, 100),
		scrollStatus:               1,
		fullScreen:                 cfg.K9s.UI.DefaultsToFullScreen,
		textWrap:                   cfg.K9s.Logger.TextWrap,
		showTime:                   cfg.K9s.Logger.ShowTime,
		shouldDisplayAllContainers: allContainers,
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
	l.Refresh()
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

// ToggleAllContainers toggles the all-containers mode.
func (l *LogIndicator) ToggleAllContainers() {
	l.allContainers = !l.allContainers
	l.Refresh()
}

func (l *LogIndicator) reset() {
	l.Clear()
	l.indicator = l.indicator[:0]
}

// Refresh updates the view.
func (l *LogIndicator) Refresh() {
	l.reset()

	var (
		toggleFmt    = "[::b]%s:["
		toggleOnFmt  = toggleFmt + string(l.styles.K9s.Views.Log.Indicator.ToggleOnColor) + "::b]On[-::] %s"
		toggleOffFmt = toggleFmt + string(l.styles.K9s.Views.Log.Indicator.ToggleOffColor) + "::d]Off[-::]%s"
	)

	if l.shouldDisplayAllContainers {
		if l.allContainers {
			l.indicator = append(l.indicator, fmt.Sprintf(toggleOnFmt, "AllContainers", spacer)...)
		} else {
			l.indicator = append(l.indicator, fmt.Sprintf(toggleOffFmt, "AllContainers", spacer)...)
		}
	}

	if l.AutoScroll() {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOnFmt, "Autoscroll", spacer)...)
	} else {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOffFmt, "Autoscroll", spacer)...)
	}

	if l.FullScreen() {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOnFmt, "FullScreen", spacer)...)
	} else {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOffFmt, "FullScreen", spacer)...)
	}

	if l.Timestamp() {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOnFmt, "Timestamps", spacer)...)
	} else {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOffFmt, "Timestamps", spacer)...)
	}

	if l.TextWrap() {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOnFmt, "Wrap", "")...)
	} else {
		l.indicator = append(l.indicator, fmt.Sprintf(toggleOffFmt, "Wrap", "")...)
	}

	_, _ = l.Write(l.indicator)
}
