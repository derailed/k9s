// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
)

// Namespace represents a namespace viewer.
type Namespace struct {
	ResourceViewer
}

// NewNamespace returns a new viewer.
func NewNamespace(gvr client.GVR) ResourceViewer {
	n := Namespace{
		ResourceViewer: NewBrowser(gvr),
	}
	n.GetTable().SetDecorateFn(n.decorate)
	n.GetTable().SetEnterFn(n.switchNs)
	n.AddBindKeysFn(n.bindKeys)

	return &n
}

func (n *Namespace) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyU:      ui.NewKeyAction("Use", n.useNsCmd, true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", n.GetTable().SortColCmd(statusCol, true), false),
	})
}

func (n *Namespace) switchNs(app *App, _ ui.Tabular, _ client.GVR, path string) {
	n.useNamespace(path)
	app.gotoResource("pods", "", false)
}

func (n *Namespace) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	n.useNamespace(path)

	return nil
}

func (n *Namespace) useNamespace(fqn string) {
	_, ns := client.Namespaced(fqn)
	if client.CleanseNamespace(n.App().Config.ActiveNamespace()) == ns {
		return
	}
	if err := n.App().switchNS(ns); err != nil {
		n.App().Flash().Err(err)
		return
	}
	if err := n.App().Config.SetActiveNamespace(ns); err != nil {
		n.App().Flash().Err(err)
		return
	}
}

func (n *Namespace) decorate(td *render.TableData) {
	if n.App().Conn() == nil || len(td.RowEvents) == 0 {
		return
	}

	// checks if all ns is in the list if not add it.
	if _, ok := td.RowEvents.FindIndex(client.NamespaceAll); !ok {
		td.RowEvents = append(td.RowEvents,
			render.RowEvent{
				Kind: render.EventUnchanged,
				Row: render.Row{
					ID:     client.NamespaceAll,
					Fields: render.Fields{client.NamespaceAll, "Active", "", "", ""},
				},
			},
		)
	}

	for _, re := range td.RowEvents {
		if data.InList(n.App().Config.FavNamespaces(), re.Row.ID) {
			re.Row.Fields[0] += favNSIndicator
			re.Kind = render.EventUnchanged
		}
		if n.App().Config.ActiveNamespace() == re.Row.ID {
			re.Row.Fields[0] += defaultNSIndicator
			re.Kind = render.EventUnchanged
		}
	}
}
