package views

import (
	"strings"

	"github.com/derailed/k9s/config"
	"github.com/derailed/k9s/resource"
	"github.com/derailed/k9s/resource/k8s"
	"github.com/gdamore/tcell"
)

type contextView struct {
	*resourceView
}

func newContextView(t string, app *appView, list resource.List, c colorerFn) resourceViewer {
	v := contextView{newResourceView(t, app, list, c).(*resourceView)}
	v.extraActionsFn = v.extraActions

	v.switchPage("ctx")

	return &v
}

func (v *contextView) useContext(*tcell.EventKey) {
	if !v.rowSelected() {
		return
	}

	ctx := strings.TrimSpace(v.selectedItem)
	if strings.HasSuffix(ctx, "*") {
		ctx = strings.TrimRight(ctx, "*")
	}
	if strings.HasSuffix(ctx, "(𝜟)") {
		ctx = strings.TrimRight(ctx, "(𝜟)")
	}

	err := v.list.Resource().(*resource.Context).Switch(ctx)
	if err != nil {
		v.app.flash(flashWarn, err.Error())
		return
	}

	config.Root.SetActiveCluster(ctx)
	config.Root.Save(k8s.ClusterInfo{})

	v.app.flash(flashInfo, "Switching context to", ctx)
	v.refresh()
	if tv, ok := v.GetPrimitive("ctx").(*tableView); ok {
		tv.table.Select(0, 0)
	}
}

func (v *contextView) extraActions(aa keyActions) {
	aa[KeyU] = keyAction{description: "Use", action: v.useContext}
}
