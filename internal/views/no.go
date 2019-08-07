package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type nodeView struct {
	*resourceView
}

func newNodeView(t string, app *appView, list resource.List) resourceViewer {
	v := nodeView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *nodeView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", v.sortColCmd(7, false), true)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", v.sortColCmd(8, false), true)
	aa[ui.KeyAltC] = ui.NewKeyAction("Sort CPU%", v.sortColCmd(9, false), true)
	aa[ui.KeyAltM] = ui.NewKeyAction("Sort MEM%", v.sortColCmd(10, false), true)
}

func (v *nodeView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, asc)
		t.Refresh()

		return nil
	}
}

func (v *nodeView) showPods(app *appView, _, res, sel string) {
	showPods(app, "", "", "spec.nodeName="+sel, v.backCmd)
}

func (v *nodeView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}

func showPods(app *appView, ns, labelSel, fieldSel string, a ui.ActionHandler) {
	list := resource.NewPodList(app.Conn(), ns)
	list.SetLabelSelector(labelSel)
	list.SetFieldSelector(fieldSel)

	pv := newPodView("Pods", app, list)
	pv.setColorerFn(podColorer)
	pv.setExtraActionsFn(func(aa ui.KeyActions) {
		aa[tcell.KeyEsc] = ui.NewKeyAction("Back", a, true)
	})
	// Reset active namespace to all.
	app.Config.SetActiveNamespace(ns)
	app.inject(pv)
}
