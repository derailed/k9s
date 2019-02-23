package views

import (
	"fmt"
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

const (
	favNSIndicator     = "+"
	defaultNSIndicator = "(*)"
	deltaNSIndicator   = "(𝜟)"
)

var nsCleanser = regexp.MustCompile(`(\w+)[+|(*)|(𝜟)]*`)

type namespaceView struct {
	*resourceView
}

func newNamespaceView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := namespaceView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.selectedFn = v.getSelectedItem
	v.decorateDataFn = v.decorate
	v.switchPage("ns")
	return &v
}

func (v *namespaceView) extraActions(aa keyActions) {
	aa[KeyU] = newKeyAction("Use", v.useNamespace)
}

func (v *namespaceView) useNamespace(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	ns := v.getSelectedItem()
	config.Root.SetActiveNamespace(ns)
	config.Root.Save()
	v.refresh()
	v.app.flash(flashInfo, fmt.Sprintf("Setting namespace `%s as your default namespace", ns))
	return nil
}

func (v *namespaceView) getSelectedItem() string {
	return v.cleanser(v.selectedItem)
}

func (*namespaceView) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
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
