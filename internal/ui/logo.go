// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
)

// Logo represents a K9s logo.
type Logo struct {
	*tview.Flex

	logo, status *tview.TextView
	lines        []string
	styles       *config.Styles
	items        int
	mx           sync.Mutex
}

// NewLogo returns a new logo.
func NewLogo(styles *config.Styles) *Logo {
	l := Logo{
		Flex:   tview.NewFlex(),
		logo:   logo(),
		status: status(),
		lines:  slices.Clone(LogoSmall),
		styles: styles,
	}
	l.SetDirection(tview.FlexRow)
	l.resize()
	l.refreshLogo(styles.Body().LogoColor)
	l.SetBackgroundColor(styles.BgColor())
	styles.AddListener(&l)

	return &l
}

// Logo returns the logo viewer.
func (l *Logo) Logo() *tview.TextView {
	return l.logo
}

// Status returns the status viewer.
func (l *Logo) Status() *tview.TextView {
	return l.status
}

// SetLogo updates the logo art.
func (l *Logo) SetLogo(art string) {
	l.mx.Lock()
	l.lines = logoLines(art)
	l.mx.Unlock()

	l.resize()
	l.refreshLogo(l.styles.Body().LogoColor)
}

// Width returns the logo preferred width.
func (l *Logo) Width() int {
	l.mx.Lock()
	defer l.mx.Unlock()

	w := 0
	for _, line := range l.lines {
		w = max(w, runewidth.StringWidth(line))
	}

	return w
}

// Height returns the logo preferred height including status.
func (l *Logo) Height() int {
	l.mx.Lock()
	defer l.mx.Unlock()

	return len(l.lines) + 1
}

// StylesChanged notifies the skin changed.
func (l *Logo) StylesChanged(s *config.Styles) {
	l.styles = s
	l.SetBackgroundColor(l.styles.BgColor())
	l.status.SetBackgroundColor(l.styles.BgColor())
	l.logo.SetBackgroundColor(l.styles.BgColor())
	l.refreshLogo(l.styles.Body().LogoColor)
}

// IsBenchmarking checks if benchmarking is active or not.
func (l *Logo) IsBenchmarking() bool {
	txt := l.Status().GetText(true)
	return strings.Contains(txt, "Bench")
}

// Reset clears out the logo view and resets colors.
func (l *Logo) Reset() {
	l.status.Clear()
	l.StylesChanged(l.styles)
}

// Err displays a log error state.
func (l *Logo) Err(msg string) {
	l.update(msg, l.styles.Body().LogoColorError)
}

// Warn displays a log warning state.
func (l *Logo) Warn(msg string) {
	l.update(msg, l.styles.Body().LogoColorWarn)
}

// Info displays a log info state.
func (l *Logo) Info(msg string) {
	l.update(msg, l.styles.Body().LogoColorInfo)
}

func (l *Logo) update(msg string, c config.Color) {
	l.refreshStatus(msg, c)
	l.refreshLogo(c)
}

func (l *Logo) refreshStatus(msg string, c config.Color) {
	l.mx.Lock()
	defer l.mx.Unlock()

	l.status.SetBackgroundColor(c.Color())
	l.status.SetText(
		fmt.Sprintf("[%s::b]%s", l.styles.Body().LogoColorMsg, msg),
	)
}

func (l *Logo) refreshLogo(c config.Color) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.logo.Clear()
	for i, s := range l.lines {
		_, _ = fmt.Fprintf(l.logo, "[%s::b]%s", c, s)
		if i+1 < len(l.lines) {
			_, _ = fmt.Fprintf(l.logo, "\n")
		}
	}
}

func logoLines(art string) []string {
	art = strings.ReplaceAll(art, "\r\n", "\n")
	art = strings.TrimRight(art, "\n")
	if strings.TrimSpace(art) == "" {
		return slices.Clone(LogoSmall)
	}

	return strings.Split(art, "\n")
}

func (l *Logo) resize() {
	for i := 0; i < l.items; i++ {
		l.RemoveItemAtIndex(0)
	}
	l.AddItem(l.logo, max(1, len(l.lines)), 1, false)
	l.AddItem(l.status, 1, 1, false)
	l.items = 2
}

func logo() *tview.TextView {
	v := tview.NewTextView()
	v.SetWordWrap(false)
	v.SetWrap(false)
	v.SetTextAlign(tview.AlignLeft)
	v.SetDynamicColors(true)

	return v
}

func status() *tview.TextView {
	v := tview.NewTextView()
	v.SetWordWrap(false)
	v.SetWrap(false)
	v.SetTextAlign(tview.AlignCenter)
	v.SetDynamicColors(true)

	return v
}
