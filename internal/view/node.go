package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
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

func (n *Node) showPods(app *App, ns, res, sel string) {
	showPods(app, n.GetTable().GetSelectedItem(), "", "spec.nodeName="+sel)
}

func (n *Node) viewCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !n.GetTable().RowSelected() {
		return evt
	}

	sel := n.GetTable().GetSelectedItem()
	log.Debug().Msgf("------ VIEW NODE %q", sel)
	o, err := n.App().factory.Client().DynDialOrDie().Resource(client.GVR(n.GVR()).AsGVR()).Get(sel, metav1.GetOptions{})
	if err != nil {
		n.App().Flash().Errf("Unable to get resource %q -- %s", n.GVR(), err)
		return nil
	}

	raw, err := toYAML(o)
	if err != nil {
		n.App().Flash().Errf("Unable to marshal resource %s", err)
		return nil
	}

	details := NewDetails("YAML")
	details.SetSubject(sel)
	details.SetTextColor(n.App().Styles.FgColor())
	details.SetText(colorizeYAML(n.App().Styles.Views().Yaml, raw))
	details.ScrollToBeginning()
	if err := n.App().inject(details); err != nil {
		n.App().Flash().Err(err)
	}

	return nil
}
