package views

import (
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
)

type contextView struct {
	*resourceView
}

func newContextView(t string, app *appView, list resource.List) resourceViewer {
	v := contextView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.getTV().cleanseFn = v.cleanser

	return &v
}

func (v *contextView) extraActions(aa keyActions) {
	delete(v.getTV().actions, KeyShiftA)
	aa[tcell.KeyEnter] = newKeyAction("Switch", v.useCmd, true)
}

func (v *contextView) useCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}
	if err := v.useContext(v.selectedItem); err != nil {
		v.app.flash().err(err)
		return evt
	}

	v.app.gotoResource("po", true)

	return nil
}

func (*contextView) cleanser(s string) string {
	name := strings.TrimSpace(s)
	if strings.HasSuffix(name, "*") {
		name = strings.TrimRight(name, "*")
	}
	if strings.HasSuffix(name, "(𝜟)") {
		name = strings.TrimRight(name, "(𝜟)")
	}
	return name
}

func (v *contextView) useContext(name string) error {
	ctx := v.cleanser(name)
	if err := v.list.Resource().(*resource.Context).Switch(ctx); err != nil {
		return err
	}

	v.app.startInformer()
	v.app.config.Reset()
	v.app.config.Save()
	v.app.stopForwarders()
	v.app.flash().infof("Switching context to %s", ctx)
	v.refresh()
	if tv, ok := v.GetPrimitive("ctx").(*tableView); ok {
		tv.Select(0, 0)
	}

	return nil
}
