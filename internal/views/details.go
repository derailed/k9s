package views

import (
	"fmt"

	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const detailFmt = " [aqua::-]%s [fuchsia::b]%s "

// detailsView display yaml output
type detailsView struct {
	*tview.TextView

	actions  keyActions
	category string
}

func newDetailsView() *detailsView {
	v := detailsView{TextView: tview.NewTextView()}
	v.TextView.SetDynamicColors(true)
	v.TextView.SetBorder(true)
	v.SetTitleColor(tcell.ColorAqua)
	v.SetInputCapture(v.keyboard)
	return &v
}

func (v *detailsView) setCategory(n string) {
	v.category = n
}

func (v *detailsView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if evt.Key() == tcell.KeyRune {
		if a, ok := v.actions[evt.Key()]; ok {
			a.action(evt)
			evt = nil
		}
	} else {
		if a, ok := v.actions[evt.Key()]; ok {
			a.action(evt)
			evt = nil
		}
	}
	return evt
}

// SetActions to handle keyboard inputs
func (v *detailsView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints fetch mmemonic and hints
func (v *detailsView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *detailsView) setTitle(t string) {
	v.SetTitle(fmt.Sprintf(detailFmt, t, v.category))
}
