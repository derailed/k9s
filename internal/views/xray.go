package views

import (
	"context"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

type xrayView struct {
	*tview.Table

	app     *appView
	actions keyActions
}

func newXrayView(app *appView) *xrayView {
	v := xrayView{app: app, Table: tview.NewTable()}
	v.SetBorder(true)
	v.SetTitle(" Details ")
	v.SetTitleColor(tcell.ColorAqua)
	v.SetSelectable(true, false)
	v.SetSelectedStyle(tcell.ColorBlack, tcell.ColorAqua, tcell.AttrNone)
	v.SetInputCapture(v.keyboard)
	return &v
}

func (v *xrayView) setTitle(t string) {
	v.Table.SetTitle(t)
}

func (v *xrayView) clear() {
	v.Table.Clear()
}

func (v *xrayView) blur() {
	v.Table.Blur()
}

func (v *xrayView) init(_ context.Context) {
}

// SetActions to handle keyboard inputs
func (v *xrayView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints fetch mmemonic and hints
func (v *xrayView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *xrayView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
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

func (v *xrayView) update(pp resource.Properties) {
	v.Clear()

	var row int
	for col, h := range pp["Headers"].(resource.Row) {
		tc := tview.NewTableCell(h)
		tc.SetExpansion(2)
		v.SetCell(0, col, tc)
	}
	row++

	for _, r := range pp["Rows"].([]resource.Row) {
		for col, c := range r {
			tc := tview.NewTableCell(c)
			tc.SetExpansion(2)
			v.SetCell(row, col, tc)
		}
		row++
	}
}
