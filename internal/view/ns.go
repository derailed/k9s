// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
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

func (n *Namespace) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
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
	if n.App().Conn() == nil || td.RowEvents.Empty() {
		return
	}
	// checks if all ns is in the list if not add it.
	if _, ok := td.RowEvents.FindIndex(client.NamespaceAll); !ok {
		td.RowEvents.Add(render.RowEvent{
			Kind: render.EventUnchanged,
			Row: render.Row{
				ID:     client.NamespaceAll,
				Fields: render.Fields{client.NamespaceAll, "Active", "", "", ""},
			},
		},
		)
	}

	favs := make(map[string]struct{})
	for _, ns := range n.App().Config.FavNamespaces() {
		favs[ns] = struct{}{}
	}
	ans := n.App().Config.ActiveNamespace()
	td.RowEvents.Range(func(i int, re render.RowEvent) bool {
		_, n := client.Namespaced(re.Row.ID)
		if _, ok := favs[n]; ok {
			re.Row.Fields[0] += favNSIndicator
		}
		if ans == re.Row.ID {
			re.Row.Fields[0] += defaultNSIndicator
		}
		re.Kind = render.EventUnchanged
		td.RowEvents.Set(i, re)
		return true
	})
}
