package view

import (
	"context"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const nodeTitle = "Nodes"

// Node represents a node view.
type Node struct {
	ResourceViewer
}

// NewNode returns a new node view.
func NewNode(title, gvr string, list resource.List) ResourceViewer {
	return &Node{
		ResourceViewer: NewResource(nodeTitle, gvr, list),
	}
}

func (n *Node) Init(ctx context.Context) error {
	if err := n.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	n.bindKeys()
	n.GetTable().SetEnterFn(n.showPods)

	return nil
}

func (n *Node) bindKeys() {
	n.Actions().Delete(ui.KeySpace, tcell.KeyCtrlSpace)
	n.Actions().Add(ui.KeyActions{
		ui.KeyShiftC: ui.NewKeyAction("Sort CPU", n.GetTable().SortColCmd(7, false), false),
		ui.KeyShiftM: ui.NewKeyAction("Sort MEM", n.GetTable().SortColCmd(8, false), false),
		ui.KeyShiftX: ui.NewKeyAction("Sort CPU%", n.GetTable().SortColCmd(9, false), false),
		ui.KeyShiftZ: ui.NewKeyAction("Sort MEM%", n.GetTable().SortColCmd(10, false), false),
	})
}

func (n *Node) showPods(app *App, ns, res, sel string) {
	showPods(app, n.GetTable().GetSelectedItem(), "", "spec.nodeName="+sel)
}

func showPods(app *App, path, labelSel, fieldSel string) {
	log.Debug().Msgf("NODE show pods %q -- %q -- %q", path, labelSel, fieldSel)
	app.switchNS("")

	list := resource.NewPodList(app.Conn(), "")
	list.SetLabelSelector(labelSel)
	list.SetFieldSelector(fieldSel)

	v := NewPod(path, "v1/pods", list)
	v.GetTable().SetColorerFn(render.Pod{}.ColorerFunc())

	ns, _ := namespaced(path)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		log.Error().Err(err).Msg("Config NS set failed!")
	}
	app.inject(v)
}
