package views

import (
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

func newNamespaceView(t string, app *appView, list resource.List) resourceViewer {
	v := namespaceView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.selectedFn = v.getSelectedItem
	v.decorateFn = v.decorate
	v.getTV().cleanseFn = v.cleanser

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
	if err := v.app.config.SetActiveNamespace(name); err != nil {
		v.app.flash().err(err)
	} else {
		v.app.flash().infof("Namespace %s is now active!", name)
	}
	v.app.config.Save()
}

func (v *namespaceView) getSelectedItem() string {
	return v.cleanser(v.selectedItem)
}

func (*namespaceView) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (v *namespaceView) decorate(data resource.TableData) resource.TableData {
	if _, ok := data.Rows[resource.AllNamespaces]; !ok {
		if v.app.conn().CanIAccess("", "namespaces", "namespace.v1", []string{"list"}) {
			data.Rows[resource.AllNamespace] = &resource.RowEvent{
				Action: resource.Unchanged,
				Fields: resource.Row{resource.AllNamespace, "Active", "0"},
				Deltas: resource.Row{"", "", ""},
			}
		}
	}
	for k, r := range data.Rows {
		if config.InList(v.app.config.FavNamespaces(), k) {
			r.Fields[0] += "+"
			r.Action = resource.Unchanged
		}
		if v.app.config.ActiveNamespace() == k {
			r.Fields[0] += "(*)"
			r.Action = resource.Unchanged
		}
	}

	return data
}
