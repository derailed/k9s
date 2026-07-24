// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// TabBar renders the one-line horizontal strip of tab labels shown above the
// content area whenever more than one tab is open.
type TabBar struct {
	*tview.TextView
	styles *config.Styles
}

// NewTabBar returns an initialised TabBar attached to the given style set.
func NewTabBar(styles *config.Styles) *TabBar {
	t := &TabBar{
		styles:   styles,
		TextView: tview.NewTextView(),
	}
	t.SetDynamicColors(true)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetBackgroundColor(styles.BgColor())
	styles.AddListener(t)
	return t
}

// StylesChanged implements config.StyleListener.
func (t *TabBar) StylesChanged(s *config.Styles) {
	t.styles = s
	t.SetBackgroundColor(s.BgColor())
}

// Render redraws the tab strip. activeIdx is the zero-based index of the
// currently active tab; labels contains one entry per open tab.
func (t *TabBar) Render(labels []string, activeIdx int) {
	t.Clear()
	var sb strings.Builder
	for i, label := range labels {
		if i == activeIdx {
			fmt.Fprintf(&sb, "[black:white:b] %d:%s [-:-:-]  ", i+1, label)
		} else {
			fmt.Fprintf(&sb, "[gray:-:-] %d:%s [-:-:-]  ", i+1, label)
		}
	}
	fmt.Fprint(t, sb.String())
}
