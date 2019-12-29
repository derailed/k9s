package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// Logo represents a K9s logo.
type Logo struct {
	*tview.Flex

	logo, status *tview.TextView
	styles       *config.Styles
}

// NewLogo returns a new logo.
func NewLogo(styles *config.Styles) *Logo {
	l := Logo{
		Flex:   tview.NewFlex(),
		logo:   logo(),
		status: status(),
		styles: styles,
	}
	l.SetDirection(tview.FlexRow)
	l.AddItem(l.logo, 0, 6, false)
	l.AddItem(l.status, 0, 1, false)
	l.refreshLogo(styles.Body().LogoColor)
	styles.AddListener(&l)

	return &l
}

func (l *Logo) Logo() *tview.TextView {
	return l.logo
}

func (l *Logo) Status() *tview.TextView {
	return l.status
}

func (l *Logo) StylesChanged(s *config.Styles) {
	l.styles = s
	l.Reset()
}

// Reset clears out the logo view and resets colors.
func (l *Logo) Reset() {
	l.status.Clear()
	l.SetBackgroundColor(l.styles.BgColor())
	l.status.SetBackgroundColor(l.styles.BgColor())
	l.logo.SetBackgroundColor(l.styles.BgColor())
	l.refreshLogo(l.styles.Body().LogoColor)
}

// Err displays a log error state.
func (l *Logo) Err(msg string) {
	l.update(msg, "red")
}

// Warn displays a log warning state.
func (l *Logo) Warn(msg string) {
	l.update(msg, "mediumvioletred")
}

// Info displays a log info state.
func (l *Logo) Info(msg string) {
	l.update(msg, "green")
}

func (l *Logo) update(msg, c string) {
	l.refreshStatus(msg, c)
	l.refreshLogo(c)
}

func (l *Logo) refreshStatus(msg, c string) {
	l.status.SetBackgroundColor(config.AsColor(c))
	l.status.SetText(fmt.Sprintf("[white::b]%s", msg))
}

func (l *Logo) refreshLogo(c string) {
	l.logo.Clear()
	for i, s := range LogoSmall {
		fmt.Fprintf(l.logo, "[%s::b]%s", c, s)
		if i+1 < len(LogoSmall) {
			fmt.Fprintf(l.logo, "\n")
		}
	}
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
