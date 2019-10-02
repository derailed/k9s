package views

import (
	"regexp"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
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

func newNamespaceView(title, gvr string, app *appView, list resource.List) resourceViewer {
	v := namespaceView{newResourceView(title, gvr, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.masterPage().SetSelectedFn(v.cleanser)
	v.decorateFn = v.decorate
	v.enterFn = v.switchNs

	return &v
}

func (v *namespaceView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyU] = ui.NewKeyAction("Use", v.useNsCmd, true)
}

func (v *namespaceView) switchNs(app *appView, _, res, sel string) {
	v.useNamespace(sel)
	app.gotoResource("po", true)
}

func (v *namespaceView) useNsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}
	v.useNamespace(v.masterPage().GetSelectedItem())

	return nil
}

func (v *namespaceView) useNamespace(ns string) {
	if err := v.app.Config.SetActiveNamespace(ns); err != nil {
		v.app.Flash().Err(err)
	} else {
		v.app.Flash().Infof("Namespace %s is now active!", ns)
	}
	v.app.Config.Save()
	v.app.startInformer(ns)
}

func (*namespaceView) cleanser(s string) string {
	return nsCleanser.ReplaceAllString(s, `$1`)
}

func (v *namespaceView) decorate(data resource.TableData) resource.TableData {
	if _, ok := data.Rows[resource.AllNamespaces]; !ok {
		if err := v.app.Conn().CheckNSAccess(""); err == nil {
			data.Rows[resource.AllNamespace] = &resource.RowEvent{
				Action: resource.Unchanged,
				Fields: resource.Row{resource.AllNamespace, "Active", "0"},
				Deltas: resource.Row{"", "", ""},
			}
		}
	}
	for k, r := range data.Rows {
		if config.InList(v.app.Config.FavNamespaces(), k) {
			r.Fields[0] += "+"
			r.Action = resource.Unchanged
		}
		if v.app.Config.ActiveNamespace() == k {
			r.Fields[0] += "(*)"
			r.Action = resource.Unchanged
		}
	}

	return data
}
