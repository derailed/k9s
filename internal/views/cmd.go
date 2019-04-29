package views

import (
	"fmt"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/tview"
)

const defaultPrompt = "%c> %s"

type cmdView struct {
	*tview.TextView

	activated bool
	icon      rune
	text      string
	app       *appView
}

func newCmdView(app *appView, ic rune) *cmdView {
	v := cmdView{app: app, icon: ic, TextView: tview.NewTextView()}
	{
		v.SetWordWrap(true)
		v.SetWrap(true)
		v.SetDynamicColors(true)
		v.SetBorder(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetBackgroundColor(app.styles.BgColor())
		v.SetBorderColor(config.AsColor(app.styles.Style.Border.FocusColor))
		v.SetTextColor(app.styles.FgColor())
	}
	return &v
}

func (v *cmdView) inCmdMode() bool {
	return v.activated
}

func (v *cmdView) activate() {
	v.write(v.text)
}

func (v *cmdView) update(s string) {
	v.text = s
	v.Clear()
	v.write(s)
}

func (v *cmdView) append(r rune) {
	fmt.Fprintf(v, string(r))
}

func (v *cmdView) write(s string) {
	fmt.Fprintf(v, defaultPrompt, v.icon, s)
}

// ----------------------------------------------------------------------------
// Event Listener protocol...

func (v *cmdView) changed(s string) {
	v.update(s)
}

func (v *cmdView) active(f bool) {
	v.activated = f
	if f {
		v.SetBorder(true)
		v.SetTextColor(v.app.styles.FgColor())
		v.activate()
	} else {
		v.SetBorder(false)
		v.SetBackgroundColor(v.app.styles.BgColor())
		v.Clear()
	}
}
