package view

import (
	"context"
	"regexp"

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
	ResourceViewer
}

// NewNamespace returns a new viewer
func NewNamespace(title, gvr string, list resource.List) ResourceViewer {
	return &Namespace{
		ResourceViewer: NewResource(title, gvr, list),
	}
}

func (n *Namespace) Init(ctx context.Context) error {
	n.GetTable().SetDecorateFn(n.decorate)
	n.GetTable().SetEnterFn(n.switchNs)
	if err := n.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	n.GetTable().SetSelectedFn(n.cleanser)
	n.bindKeys()

	return nil
}

func (n *Namespace) bindKeys() {
	n.Actions().Add(ui.KeyActions{
		ui.KeyU: ui.NewKeyAction("Use", n.useNsCmd, true),
	})
}

func (n *Namespace) switchNs(app *App, _, res, sel string) {
	n.useNamespace(sel)
	app.gotoResource("po")
}

func (n *Namespace) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	ns := n.GetTable().GetSelectedItem()
	if ns == "" {
		return evt
	}
	n.useNamespace(ns)

	return nil
}

func (n *Namespace) useNamespace(ns string) {
	if err := n.App().Config.SetActiveNamespace(ns); err != nil {
		n.App().Flash().Err(err)
	} else {
		n.App().Flash().Infof("Namespace %s is now active!", ns)
	}
	if err := n.App().Config.Save(); err != nil {
		log.Error().Err(err).Msg("Config file save failed!")
	}
	n.App().switchNS(ns)
}

func (*Namespace) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (n *Namespace) decorate(data resource.TableData) resource.TableData {
	return resource.TableData{}
	// BOZO!!
	// if n.App().Conn() == nil {
	// 	return resource.TableData{}
	// }

	// if _, ok := data.Rows[resource.AllNamespaces]; !ok {
	// 	if err := n.App().Conn().CheckNSAccess(""); err == nil {
	// 		data.Rows[resource.AllNamespace] = &resource.RowEvent{
	// 			Action: resource.Unchanged,
	// 			Fields: resource.Row{resource.AllNamespace, "Active", "0"},
	// 			Deltas: resource.Row{"", "", ""},
	// 		}
	// 	}
	// }
	// for k, r := range data.Rows {
	// 	if config.InList(n.App().Config.FavNamespaces(), k) {
	// 		r.Fields[0] += favNSIndicator
	// 		r.Action = resource.Unchanged
	// 	}
	// 	if n.App().Config.ActiveNamespace() == k {
	// 		r.Fields[0] += defaultNSIndicator
	// 		r.Action = resource.Unchanged
	// 	}
	// }

	// return data
}
