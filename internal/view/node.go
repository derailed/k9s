package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Node represents a node view.
type Node struct {
	ResourceViewer
}

// NewNode returns a new node view.
func NewNode(gvr client.GVR) ResourceViewer {
	n := Node{
		ResourceViewer: NewBrowser(gvr),
	}
	n.SetBindKeysFn(n.bindKeys)
	n.GetTable().SetEnterFn(n.showPods)

	return &n
}

func (n *Node) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeySpace, tcell.KeyCtrlSpace, tcell.KeyCtrlD)
	aa.Add(ui.KeyActions{
		ui.KeyY:      ui.NewKeyAction("YAML", n.viewCmd, true),
		ui.KeyShiftC: ui.NewKeyAction("Sort CPU", n.GetTable().SortColCmd(7, false), false),
		ui.KeyShiftM: ui.NewKeyAction("Sort MEM", n.GetTable().SortColCmd(8, false), false),
		ui.KeyShiftX: ui.NewKeyAction("Sort CPU%", n.GetTable().SortColCmd(9, false), false),
		ui.KeyShiftZ: ui.NewKeyAction("Sort MEM%", n.GetTable().SortColCmd(10, false), false),
	})
}

func (n *Node) showPods(app *App, _ ui.Tabular, _, path string) {
	showPods(app, n.GetTable().GetSelectedItem(), "", "spec.nodeName="+path)
}

func (n *Node) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	sel := n.GetTable().GetSelectedItem()
	gvr := client.NewGVR(n.GVR()).GVR()
	o, err := n.App().factory.Client().DynDialOrDie().Resource(gvr).Get(sel, metav1.GetOptions{})
	if err != nil {
		n.App().Flash().Errf("Unable to get resource %q -- %s", n.GVR(), err)
		return nil
	}

	raw, err := dao.ToYAML(o)
	if err != nil {
		n.App().Flash().Errf("Unable to marshal resource %s", err)
		return nil
	}

	details := NewDetails(n.App(), "YAML", sel, true).Update(raw)
	if err := n.App().inject(details); err != nil {
		n.App().Flash().Err(err)
	}

	return nil
}
