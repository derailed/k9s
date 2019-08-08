package views

import (
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/rs/zerolog/log"
)

type contextView struct {
	*resourceView
}

func newContextView(title string, app *appView, list resource.List) resourceViewer {
	v := contextView{newResourceView(title, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions
	v.enterFn = v.useCtx
	v.masterPage().SetSelectedFn(v.cleanser)

	return &v
}

func (v *contextView) extraActions(aa ui.KeyActions) {
	v.masterPage().RmAction(ui.KeyShiftA)
}

func (v *contextView) useCtx(app *appView, _, res, sel string) {
	if err := v.useContext(sel); err != nil {
		app.Flash().Err(err)
		return
	}
	app.gotoResource("po", true)
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

	v.app.stopForwarders()
	ns, err := v.app.Conn().Config().CurrentNamespaceName()
	if err != nil {
		log.Info().Err(err).Msg("No namespace specified using all namespaces")
	}
	v.app.startInformer(ns)
	v.app.Config.Reset()
	v.app.Config.Save()
	v.app.Flash().Infof("Switching context to %s", ctx)
	v.refresh()
	if tv, ok := v.GetPrimitive("ctx").(*tableView); ok {
		tv.Select(0, 0)
	}

	return nil
}
