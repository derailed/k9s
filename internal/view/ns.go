package view

import (
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
	deltaNSIndicator   = "(𝜟)"
)

var nsCleanser = regexp.MustCompile(`(\w+)[+|(*)|(𝜟)]*`)

// Namespace represents a namespace viewer.
type Namespace struct {
	*Resource
}

// NewNamespace returns a new viewer
func NewNamespace(title, gvr string, list resource.List) ResourceViewer {
	n := Namespace{
		Resource: NewResource(title, gvr, list),
	}
	n.extraActionsFn = n.extraActions
	n.masterPage().SetSelectedFn(n.cleanser)
	n.decorateFn = n.decorate
	n.enterFn = n.switchNs

	return &n
}

func (n *Namespace) extraActions(aa ui.KeyActions) {
	aa[ui.KeyU] = ui.NewKeyAction("Use", n.useNsCmd, true)
}

func (n *Namespace) switchNs(app *App, _, res, sel string) {
	n.useNamespace(sel)
	app.gotoResource("po", true)
}

func (n *Namespace) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !n.masterPage().RowSelected() {
		return evt
	}
	n.useNamespace(n.masterPage().GetSelectedItem())

	return nil
}

func (n *Namespace) useNamespace(ns string) {
	if err := n.app.Config.SetActiveNamespace(ns); err != nil {
		n.app.Flash().Err(err)
	} else {
		n.app.Flash().Infof("Namespace %s is now active!", ns)
	}
	n.app.Config.Save()
	n.app.startInformer(ns)
}

func (*Namespace) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (n *Namespace) decorate(data resource.TableData) resource.TableData {
	if _, ok := data.Rows[resource.AllNamespaces]; !ok {
		if err := n.app.Conn().CheckNSAccess(""); err == nil {
			data.Rows[resource.AllNamespace] = &resource.RowEvent{
				Action: resource.Unchanged,
				Fields: resource.Row{resource.AllNamespace, "Active", "0"},
				Deltas: resource.Row{"", "", ""},
			}
		}
	}
	for k, r := range data.Rows {
		if config.InList(n.app.Config.FavNamespaces(), k) {
			r.Fields[0] += "+"
			r.Action = resource.Unchanged
		}
		if n.app.Config.ActiveNamespace() == k {
			r.Fields[0] += "(*)"
			r.Action = resource.Unchanged
		}
	}

	return data
}
