package view

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Node represents a node view.
type Node struct {
	*Resource
}

// NewNode returns a new node view.
func NewNode(title, gvr string, list resource.List) ResourceViewer {
	n := Node{
		Resource: NewResource(title, gvr, list),
	}
	n.extraActionsFn = n.extraActions
	n.enterFn = n.showPods

	return &n
}

func (n *Node) extraActions(aa ui.KeyActions) {
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort CPU", n.sortColCmd(7, false), false)
	aa[ui.KeyShiftM] = ui.NewKeyAction("Sort MEM", n.sortColCmd(8, false), false)
	aa[ui.KeyShiftX] = ui.NewKeyAction("Sort CPU%", n.sortColCmd(9, false), false)
	aa[ui.KeyShiftZ] = ui.NewKeyAction("Sort MEM%", n.sortColCmd(10, false), false)
}

func (n *Node) showPods(app *App, _, _, sel string) {
	showPods(app, "", "", "spec.nodeName="+sel, n.backCmd)
}

func (n *Node) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// BOZO!!
	// n.App.inject(v)

	return nil
}

func showPods(app *App, ns, labelSel, fieldSel string, a ui.ActionHandler) {
	app.switchNS(ns)

	list := resource.NewPodList(app.Conn(), ns)
	list.SetLabelSelector(labelSel)
	list.SetFieldSelector(fieldSel)

	v := NewPod("Pod", "v1/pods", list)
	v.setColorerFn(podColorer)
	// BOZO!!
	// v.masterPage().AddActions(ui.KeyActions{
	// 	tcell.KeyEsc: ui.NewKeyAction("Back", a, true),
	// })
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	app.inject(v)
}
