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
	favNSIndicator           = " ❤️ "
	defaultNSIndicator       = "(*)"
	deleteNumericBindingsKey = "delete-ns-numeric-bindings"
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
		// ui.KeyShiftD: ui.NewKeyAction("Delete Binding", n.deleteNamespaceKeyBindings, true),
		ui.KeyF: ui.NewKeyAction("Toggle Favorite", n.toggleFavorite, true),
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
		_, ns := client.Namespaced(re.Row.ID)

		if data.InList(n.App().Config.FavNamespaces(), ns) {
			re.Row.Fields[0] += favNSIndicator
			re.Kind = render.EventUnchanged
		}
		if n.App().Config.ActiveNamespace() == ns {
			re.Row.Fields[0] += defaultNSIndicator
			re.Kind = render.EventUnchanged
		}
	}
}

func (n *Namespace) toggleFavorite(evt *tcell.EventKey) *tcell.EventKey {
	_, ns := client.Namespaced(n.GetTable().GetSelectedItem())

	if ns == "" {
		return nil
	}

	ctx, err := n.App().Config.K9s.ActiveContext()

	if err != nil {
		return evt
	}

	if data.InList(n.App().Config.FavNamespaces(), ns) {
		ctx.Namespace.RmFavNS(ns)
	} else {
		ctx.Namespace.AddFavNS(ns)
	}

	return nil
}

// func (n *Namespace) deleteNamespaceKeyBindings(evt *tcell.EventKey) *tcell.EventKey {
// 	app := n.App()
// 	cl, err := app.Config.K9s.ActiveContext()

// 	if err != nil {
// 		return evt
// 	}

// 	styles := app.Styles.Dialog()
// 	bindingsToDelete := map[string]struct{}{}

// 	f := tview.NewForm().
// 		SetItemPadding(0).
// 		SetButtonsAlign(tview.AlignCenter).
// 		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
// 		SetButtonTextColor(styles.ButtonFgColor.Color()).
// 		SetLabelColor(styles.LabelFgColor.Color()).
// 		SetFieldTextColor(styles.FieldFgColor.Color()).
// 		SetFieldBackgroundColor(styles.BgColor.Color())

// 	for _, namespaceName := range cl.Namespace.Favorites {
// 		// The browser already reserves Key0 to the `all` namespace
// 		// shortcut, so we only allow deletion of other favorites namespaces
// 		if client.IsAllNamespace(namespaceName) {
// 			continue
// 		}

// 		f.AddCheckbox(namespaceName, false, func(label string, _ bool) {
// 			bindingsToDelete[label] = struct{}{}
// 		})
// 	}

// 	f.AddButton("Cancel", func() {
// 		app.Content.RemovePage(deleteNumericBindingsKey)
// 	}).AddButton("Ok", func() {
// 		app.Content.RemovePage(deleteNumericBindingsKey)

// 		if len(bindingsToDelete) == 0 {
// 			return
// 		}

// 		// Rebuild list of favorite namespaces based on what was marked for deletion
// 		newFavorites := make([]string, 0)

// 		for _, favNsName := range cl.Namespace.Favorites {
// 			if _, ok := bindingsToDelete[favNsName]; ok {
// 				continue
// 			}

// 			newFavorites = append(newFavorites, favNsName)
// 		}

// 		cl.Namespace.Favorites = newFavorites
// 	})

// 	for i := 0; i < f.GetButtonCount(); i++ {
// 		if b := f.GetButton(i); b != nil {
// 			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
// 			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
// 		}
// 	}

// 	modal := tview.NewModalForm("Delete Numeric Bindings", f)
// 	app.Content.Pages.AddPage(deleteNumericBindingsKey, modal, false, true)
// 	app.Content.Pages.ShowPage(deleteNumericBindingsKey)
// 	app.SetFocus(app.Content.Pages.GetPrimitive(deleteNumericBindingsKey))

// 	return nil
// }
