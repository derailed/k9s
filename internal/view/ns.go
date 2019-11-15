package view

import (
	"context"
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
)

var nsCleanser = regexp.MustCompile(`(\w+)[+|(*)|(𝜟)]*`)

// Namespace represents a namespace viewer.
type Namespace struct {
	*Resource
}

// NewNamespace returns a new viewer
func NewNamespace(title, gvr string, list resource.List) ResourceViewer {
	return &Namespace{
		Resource: NewResource(title, gvr, list),
	}
}

func (n *Namespace) Init(ctx context.Context) {
	n.extraActionsFn = n.extraActions
	n.decorateFn = n.decorate
	n.enterFn = n.switchNs
	n.Resource.Init(ctx)
	n.masterPage().SetSelectedFn(n.cleanser)
}

func (n *Namespace) extraActions(aa ui.KeyActions) {
	aa[ui.KeyU] = ui.NewKeyAction("Use", n.useNsCmd, true)
}

func (n *Namespace) switchNs(app *App, _, res, sel string) {
	n.useNamespace(sel)
	app.gotoResource("po")
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
	if err := n.app.Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config file save failed!")
	}
	n.app.switchNS(ns)
}

func (*Namespace) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (n *Namespace) decorate(data resource.TableData) resource.TableData {
	if n.app.Conn() == nil {
		return resource.TableData{}
	}

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
			r.Fields[0] += favNSIndicator
			r.Action = resource.Unchanged
		}
		if n.app.Config.ActiveNamespace() == k {
			r.Fields[0] += defaultNSIndicator
			r.Action = resource.Unchanged
		}
	}

	return data
}
