package views

import (
	"fmt"
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/k8s"
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
	aa[tcell.KeyEnter] = newKeyAction("Switch", v.switchNsCmd, true)
	aa[KeyU] = newKeyAction("Use", v.useNsCmd, true)
}

func (v *namespaceView) switchNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.useNamespace(v.getSelectedItem())
	v.app.gotoResource("po", true)
	return nil
}

func (v *namespaceView) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	v.useNamespace(v.getSelectedItem())
	return nil
}

func (v *namespaceView) useNamespace(name string) {
	if err := config.Root.SetActiveNamespace(name); err != nil {
		v.app.flash(flashErr, err.Error())
	} else {
		v.app.flash(flashInfo, fmt.Sprintf("Namespace %s is now active!", name))
	}
	config.Root.Save()
}

func (v *namespaceView) getSelectedItem() string {
	return v.cleanser(v.selectedItem)
}

func (*namespaceView) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (v *namespaceView) decorate(data resource.TableData) resource.TableData {
	if _, ok := data.Rows[resource.AllNamespaces]; !ok {
		if k8s.CanIAccess("", "list", "namespaces", "namespace.v1") {
			data.Rows[resource.AllNamespace] = &resource.RowEvent{
				Action: resource.Unchanged,
				Fields: resource.Row{resource.AllNamespace, "Active", "0"},
				Deltas: resource.Row{"", "", ""},
			}
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
