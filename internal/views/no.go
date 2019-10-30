package views

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type nodeView struct {
	*resourceView
}

func newNodeView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := nodeView{newResourceView(title, gvr, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *nodeView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", v.sortColCmd(7, false), false)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", v.sortColCmd(8, false), false)
	aa[ui.KeyShiftX] = ui.NewKeyAction("Sort CPU%", v.sortColCmd(9, false), false)
	aa[ui.KeyShiftZ] = ui.NewKeyAction("Sort MEM%", v.sortColCmd(10, false), false)
}

func (v *nodeView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}

func (v *nodeView) showPods(app *appView, _, _, sel string) {
	showPods(app, "", "", "spec.nodeName="+sel, v.backCmd)
}

func (v *nodeView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}

func showPods(app *appView, ns, labelSel, fieldSel string, a ui.ActionHandler) {
	app.switchNS(ns)

	list := resource.NewPodList(app.Conn(), ns)
	list.SetLabelSelector(labelSel)
	list.SetFieldSelector(fieldSel)

	pv := newPodView("Pod", "v1/pods", app, list)
	pv.setColorerFn(podColorer)
	pv.masterPage().SetActions(ui.KeyActions{
		tcell.KeyEsc: ui.NewKeyAction("Back", a, true),
	})
	// Reset active namespace to ns.
	app.Config.SetActiveNamespace(ns)
	app.inject(pv)
}
