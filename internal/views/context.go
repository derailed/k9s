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
	{
		v.extraActionsFn = v.extraActions
		v.switchPage("ctx")
	}
	return &v
}

func (v *contextView) extraActions(aa keyActions) {
	aa[tcell.KeyEnter] = newKeyAction("Switch", v.useCmd, true)
}

func (v *contextView) useCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	if err := v.useContext(v.selectedItem); err != nil {
		v.app.flash(flashWarn, err.Error())
		return evt
	}

	v.app.gotoResource("po", true)

	return nil
}

func (v *contextView) useContext(name string) error {
	ctx := strings.TrimSpace(name)
	if strings.HasSuffix(ctx, "*") {
		ctx = strings.TrimRight(ctx, "*")
	}
	if strings.HasSuffix(ctx, "(𝜟)") {
		ctx = strings.TrimRight(ctx, "(𝜟)")
	}

	if err := v.list.Resource().(*resource.Context).Switch(ctx); err != nil {
		return err
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
