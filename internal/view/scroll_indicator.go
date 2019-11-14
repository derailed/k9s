package view

import (
	"fmt"
	"sync/atomic"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// AutoScrollIndicator represents a log autoscroll status indicator.
type AutoScrollIndicator struct {
	*tview.TextView

	styles       *config.Styles
	scrollStatus int32
}

// NewAutoScrollIndicator returns a new indicator.
func NewAutoScrollIndicator(styles *config.Styles) *AutoScrollIndicator {
	a := AutoScrollIndicator{
		styles:       styles,
		TextView:     tview.NewTextView(),
		scrollStatus: 1,
	}
	a.SetBackgroundColor(config.AsColor(styles.Views().Log.BgColor))
	a.SetTextAlign(tview.AlignRight)
	a.SetDynamicColors(true)

	return &a
}

func (a *AutoScrollIndicator) AutoScroll() bool {
	return atomic.LoadInt32(&a.scrollStatus) == 1
}

func (a *AutoScrollIndicator) ToggleAutoScroll() {
	var val int32 = 1
	if a.AutoScroll() {
		val = 0
	}
	atomic.StoreInt32(&a.scrollStatus, val)
}

func (a *AutoScrollIndicator) Refresh() {
	autoScroll := "Off"
	if a.AutoScroll() {
		autoScroll = "On"
	}
	a.update("Autoscroll: " + autoScroll)
}

func (a *AutoScrollIndicator) update(status string) {
	a.Clear()
	fg, bg := a.styles.Frame().Crumb.FgColor, a.styles.Frame().Crumb.ActiveColor
	fmt.Fprintf(a, "[%s:%s:b] %-15s ", fg, bg, status)
}
