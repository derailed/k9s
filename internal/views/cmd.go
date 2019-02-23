package views

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/derailed/tview"
	log "github.com/sirupsen/logrus"
)

const defaultPrompt = "%c> %s"

type cmdView struct {
	*tview.TextView

	activated bool
	icon      rune
	text      string
}

func newCmdView(ic rune) *cmdView {
	v := cmdView{icon: ic, TextView: tview.NewTextView()}
	{
		v.SetWordWrap(false)
		v.SetWrap(false)
		v.SetDynamicColors(true)
		v.SetBorderPadding(0, 0, 1, 1)
		v.SetTextColor(tcell.ColorAqua)
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
		log.Debug("CmdView was activated...")
		v.activate()
	} else {
		log.Debug("CmdView was deactivated!")
		v.Clear()
	}
}
