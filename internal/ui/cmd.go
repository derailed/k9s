package ui

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

const defaultPrompt = "%c> %s"

// CmdView captures users free from command input.
type CmdView struct {
	*tview.TextView

	activated bool
	icon      rune
	text      string
	styles    *config.Styles
}

// NewCmdView returns a new command view.
func NewCmdView(styles *config.Styles, ic rune) *CmdView {
	v := CmdView{styles: styles, icon: ic, TextView: tview.NewTextView()}
	{
		v.SetWordWrap(true)
		v.SetWrap(true)
		v.SetDynamicColors(true)
		v.SetBorder(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetBackgroundColor(styles.BgColor())
		v.SetBorderColor(config.AsColor(styles.Frame().Border.FocusColor))
		v.SetTextColor(styles.FgColor())
	}
	return &v
}

// InCmdMode returns true if command is active, false otherwise.
func (v *CmdView) InCmdMode() bool {
	return v.activated
}

func (v *CmdView) activate() {
	v.write(v.text)
}

func (v *CmdView) update(s string) {
	v.text = s
	v.Clear()
	v.write(s)
}

func (v *CmdView) append(r rune) {
	fmt.Fprintf(v, string(r))
}

func (v *CmdView) write(s string) {
	fmt.Fprintf(v, defaultPrompt, v.icon, s)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

func (v *CmdView) changed(s string) {
	v.update(s)
}

func (v *CmdView) active(f bool) {
	v.activated = f
	if f {
		v.SetBorder(true)
		v.SetTextColor(v.styles.FgColor())
		v.activate()
	} else {
		v.SetBorder(false)
		v.SetBackgroundColor(v.styles.BgColor())
		v.Clear()
	}
}
