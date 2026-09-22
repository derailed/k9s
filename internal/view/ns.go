// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	favNSIndicator     = "+"
	recentNSIndicator  = "~"
	defaultNSIndicator = "(*)"
)

// Namespace represents a namespace viewer.
type Namespace struct {
	ResourceViewer
}

// NewNamespace returns a new viewer.
func NewNamespace(gvr *client.GVR) ResourceViewer {
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
		ui.KeyU:        ui.NewKeyAction("Use", n.useNsCmd, true),
		tcell.KeyCtrlF: ui.NewKeyAction("Toggle Favorites", n.toggleFavsCmd, true),
		ui.KeyF:        ui.NewKeyAction("Fav/Unfav", n.toggleFavCmd, true),
	})
}

func (n *Namespace) toggleFavsCmd(*tcell.EventKey) *tcell.EventKey {
	n.GetTable().ToggleFavs()
	return nil
}

func (n *Namespace) toggleFavCmd(*tcell.EventKey) *tcell.EventKey {
	path := n.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	_, ns := client.Namespaced(path)
	if ns == client.NamespaceAll {
		n.App().Flash().Warn("Namespace \"all\" is always mapped to slot 0. It can't be favorited")
		return nil
	}
	if n.App().Config.IsFavNamespace(ns) {
		if err := n.App().Config.RemoveFavNamespace(ns); err != nil {
			n.App().Flash().Err(err)
			return nil
		}
		n.App().Flash().Infof("Removed %q from favorites", ns)
	} else {
		if err := n.App().Config.AddFavNamespace(ns); err != nil {
			n.App().Flash().Err(err)
			return nil
		}
		n.App().Flash().Infof("Added %q to favorites", ns)
	}
	n.Refresh()

	return nil
}

func (n *Namespace) switchNs(app *App, _ ui.Tabular, _ *client.GVR, path string) {
	n.useNamespace(path)
	_, ns := client.Namespaced(path)
	app.gotoResource(client.PodGVR.String()+" "+ns, "", false, true)
}

func (n *Namespace) useNsCmd(*tcell.EventKey) *tcell.EventKey {
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

func (n *Namespace) decorate(td *model1.TableData) {
	if n.App().Conn() == nil || td.RowCount() == 0 {
		return
	}
	// checks if all ns is in the list if not add it.
	if _, ok := td.FindRow(client.NamespaceAll); !ok {
		td.AddRow(model1.RowEvent{
			Kind: model1.EventUnchanged,
			Row: model1.Row{
				ID:     client.NamespaceAll,
				Fields: model1.Fields{client.NamespaceAll, "Active", "", "", ""},
			},
		},
		)
	}

	var (
		cfg      = n.App().Config
		all      = sets.New(cfg.FavNamespaces()...) // favorites+recent, ie slotted
		activeNS = cfg.ActiveNamespace()
	)
	td.RowsRange(func(i int, re model1.RowEvent) bool {
		_, n := client.Namespaced(re.Row.ID)
		switch {
		case n == client.NamespaceAll:
			// Always slot 0. Not part of favorites/recent.
		case cfg.IsFavNamespace(n):
			re.Row.Fields[0] += favNSIndicator
		case all.Has(n):
			re.Row.Fields[0] += recentNSIndicator
		}
		if n == activeNS {
			re.Row.Fields[0] += defaultNSIndicator
		}
		re.Kind = model1.EventUnchanged
		td.SetRow(i, re)
		return true
	})
}
