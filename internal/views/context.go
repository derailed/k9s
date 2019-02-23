package views

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
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

func (v *contextView) useContext(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
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
		return evt
	}

	config.Root.Reset()
	config.Root.Save()

	v.app.flash(flashInfo, "Switching context to", ctx)
	v.refresh()
	if tv, ok := v.GetPrimitive("ctx").(*tableView); ok {
		tv.Select(0, 0)
	}
	return nil
}

func (v *contextView) extraActions(aa keyActions) {
	aa[KeyU] = newKeyAction("Use", v.useContext)
}
