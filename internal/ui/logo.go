package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

// LogoView represents a K9s logo.
type LogoView struct {
	*tview.Flex
	logo, status *tview.TextView
	styles       *config.Styles
}

// NewLogoView returns a new logo.
func NewLogoView(styles *config.Styles) *LogoView {
	v := LogoView{
		Flex:   tview.NewFlex(),
		logo:   logo(),
		status: status(),
		styles: styles,
	}
	v.SetDirection(tview.FlexRow)
	v.AddItem(v.logo, 0, 6, false)
	v.AddItem(v.status, 0, 1, false)
	v.refreshLogo(styles.Body().LogoColor)

	return &v
}

// Reset clears out the logo view and resets colors.
func (v *LogoView) Reset() {
	v.status.Clear()
	v.status.SetBackgroundColor(v.styles.BgColor())
	v.refreshLogo(v.styles.Body().LogoColor)
}

// Err displays a log error state.
func (v *LogoView) Err(msg string) {
	v.update(msg, "red")
}

// Warn displays a log warning state.
func (v *LogoView) Warn(msg string) {
	v.update(msg, "mediumvioletred")
}

// Info displays a log info state.
func (v *LogoView) Info(msg string) {
	v.update(msg, "green")
}

func (v *LogoView) update(msg, c string) {
	v.refreshStatus(msg, c)
	v.refreshLogo(c)
}

func (v *LogoView) refreshStatus(msg, c string) {
	v.status.SetBackgroundColor(config.AsColor(c))
	v.status.SetText(fmt.Sprintf("[white::b]%s", msg))
}

func (v *LogoView) refreshLogo(c string) {
	v.logo.Clear()
	for i, s := range LogoSmall {
		fmt.Fprintf(v.logo, "[%s::b]%s", c, s)
		if i+1 < len(LogoSmall) {
			fmt.Fprintf(v.logo, "\n")
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
