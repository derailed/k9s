package views

import (
	"fmt"

	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
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

	k9sCfg.K9s.Namespace.Active = v.selectedItem
	k9sCfg.addFavNS(v.selectedItem)
	k9sCfg.validateAndSave()
	v.app.flash(flashInfo, fmt.Sprintf("Setting namespace `%s as your default namespace", v.selectedItem))
}

func (v *namespaceView) extraActions(aa keyActions) {
	aa[tcell.KeyCtrlS] = keyAction{description: "Switch", action: v.useNamespace}
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
		if k9sCfg.K9s.Namespace.Active == k {
			v.Fields[0] = v.Fields[0] + "*"
			v.Action = resource.Unchanged
		}
	}

	return data
}
