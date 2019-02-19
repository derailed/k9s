package views

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
	deltaNSIndicator   = "(𝜟)"
)

type namespaceView struct {
	*resourceView
}

func newNamespaceView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := namespaceView{
		resourceView: newResourceView(t, app, list, c).(*resourceView),
	}
	v.extraActionsFn = v.extraActions
	v.decorateDataFn = v.decorate
	v.switchPage("ns")
	return &v
}

func (v *namespaceView) useNamespace(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	ns := v.selectedItem
	for _, i := range []string{deltaNSIndicator, favNSIndicator, defaultNSIndicator} {
		if strings.HasSuffix(ns, i) {
			ns = strings.TrimRight(ns, i)
		}
	}
	v.refresh()

	config.Root.SetActiveNamespace(ns)
	config.Root.Save()
	v.app.flash(flashInfo, fmt.Sprintf("Setting namespace `%s as your default namespace", ns))
}

func (v *namespaceView) extraActions(aa keyActions) {
	aa[KeyU] = keyAction{description: "Use", action: v.useNamespace}
}

func (v *namespaceView) decorate(data resource.TableData) resource.TableData {
	if _, ok := data.Rows[resource.AllNamespaces]; !ok {
		data.Rows[resource.AllNamespace] = &resource.RowEvent{
			Action: resource.Unchanged,
			Fields: resource.Row{resource.AllNamespace, "Active", "0"},
			Deltas: resource.Row{"", "", ""},
		}
	}

	for k, v := range data.Rows {
		if config.InList(config.Root.FavNamespaces(), k) {
			v.Fields[0] += "+"
			v.Action = resource.Unchanged
		}

		if config.Root.ActiveNamespace() == k {
			v.Fields[0] += "(*)"
			v.Action = resource.Unchanged
		}
	}

	return data
}
