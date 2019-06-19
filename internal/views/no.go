package views

import (
	"github.com/derailed/k9s/internal/resource"
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

func (v *nodeView) extraActions(aa keyActions) {
	aa[KeyShiftC] = newKeyAction("Sort CPU", v.sortColCmd(7, false), true)
	aa[KeyShiftM] = newKeyAction("Sort MEM", v.sortColCmd(8, false), true)
	aa[KeyAltC] = newKeyAction("Sort CPU%", v.sortColCmd(9, false), true)
	aa[KeyAltM] = newKeyAction("Sort MEM%", v.sortColCmd(10, false), true)
}

func (v *nodeView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

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

func showPods(app *appView, ns, labelSel, fieldSel string, a actionHandler) {
	list := resource.NewPodList(app.conn(), ns)
	list.SetLabelSelector(labelSel)
	list.SetFieldSelector(fieldSel)

	pv := newPodView("Pods", app, list)
	pv.setColorerFn(podColorer)
	pv.setExtraActionsFn(func(aa keyActions) {
		aa[tcell.KeyEsc] = newKeyAction("Back", a, true)
	})
	// Reset active namespace to all.
	app.config.SetActiveNamespace(ns)
	app.inject(pv)
}
