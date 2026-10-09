// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"fmt"
	"slices"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	runewidth "github.com/mattn/go-runewidth"
)

const (
	// LogoWidth is the maximum width of header art in terminal columns.
	LogoWidth = 26
	// LogoArtHeight is the fixed number of art rows, excluding status.
	LogoArtHeight = 6
)

// Logo represents a K9s logo.
type Logo struct {
	*tview.Flex

	logo, status *tview.TextView
	lines        []string
	styles       *config.Styles
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
	l.AddItem(l.logo, LogoArtHeight, 0, false)
	l.AddItem(l.status, 1, 0, false)
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
	defer l.mx.Unlock()
	lines := logoLines(art)
	if slices.Equal(l.lines, lines) {
		return
	}
	l.lines = lines
	l.refreshLogo(l.styles.Body().LogoColor)
}

// Draw clips art to the available rows while preserving the status row.
func (l *Logo) Draw(screen tcell.Screen) {
	l.mx.Lock()
	defer l.mx.Unlock()
	_, _, _, height := l.GetInnerRect()
	l.ResizeItem(l.logo, min(LogoArtHeight, max(0, height-1)), 0)
	l.ResizeItem(l.status, min(1, max(0, height)), 0)
	l.logo.ScrollToBeginning()
	l.Flex.Draw(screen)
}

// StylesChanged notifies the skin changed.
func (l *Logo) StylesChanged(s *config.Styles) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.styles = s
	l.applyStyles()
}

func (l *Logo) applyStyles() {
	l.SetBackgroundColor(l.styles.BgColor())
	l.status.SetBackgroundColor(l.styles.BgColor())
	l.logo.SetBackgroundColor(l.styles.BgColor())
	l.refreshLogo(l.styles.Body().LogoColor)
}

// IsBenchmarking checks if benchmarking is active or not.
func (l *Logo) IsBenchmarking() bool {
	l.mx.Lock()
	defer l.mx.Unlock()
	txt := l.Status().GetText(true)
	return strings.Contains(txt, "Bench")
}

// Reset clears out the logo view and resets colors.
func (l *Logo) Reset() {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.status.Clear()
	l.applyStyles()
}

// Err displays a log error state.
func (l *Logo) Err(msg string) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.update(msg, l.styles.Body().LogoColorError)
}

// Warn displays a log warning state.
func (l *Logo) Warn(msg string) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.update(msg, l.styles.Body().LogoColorWarn)
}

// Info displays a log info state.
func (l *Logo) Info(msg string) {
	l.mx.Lock()
	defer l.mx.Unlock()
	l.update(msg, l.styles.Body().LogoColorInfo)
}

func (l *Logo) update(msg string, c config.Color) {
	l.refreshStatus(msg, c)
	l.refreshLogo(c)
}

func (l *Logo) refreshStatus(msg string, c config.Color) {
	l.status.SetBackgroundColor(c.Color())
	l.status.SetText(
		fmt.Sprintf("[%s::b]%s", l.styles.Body().LogoColorMsg, msg),
	)
}

func (l *Logo) refreshLogo(c config.Color) {
	l.logo.Clear()
	for i, s := range l.lines {
		_, _ = fmt.Fprintf(l.logo, "[%s::b]%s", c, tview.Escape(s))
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

	lines := strings.SplitN(art, "\n", LogoArtHeight+1)
	lines = lines[:min(len(lines), LogoArtHeight)]
	for i, line := range lines {
		line = strings.ReplaceAll(line, "\t", strings.Repeat(" ", tview.TabSize))
		lines[i] = runewidth.Truncate(line, LogoWidth, "")
	}
	return lines
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
