package view

import (
	"errors"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Context presents a context viewer.
type Context struct {
	ResourceViewer
}

// NewContext returns a new viewer.
func NewContext(gvr client.GVR) ResourceViewer {
	c := Context{
		ResourceViewer: NewBrowser(gvr),
	}
	c.GetTable().SetEnterFn(c.useCtx)
	c.GetTable().SetColorerFn(render.Context{}.ColorerFunc())
	c.SetBindKeysFn(c.bindKeys)

	return &c
}

func (c *Context) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
}

func (c *Context) useCtx(app *App, model ui.Tabular, gvr, path string) {
	log.Debug().Msgf("SWITCH CTX %q--%q", gvr, path)
	if err := useContext(app, path); err != nil {
		app.Flash().Err(err)
		return
	}
	c.Refresh()
	c.GetTable().Select(1, 0)
}

func useContext(app *App, name string) error {
	if app.Content.Top() != nil {
		app.Content.Top().Stop()
	}
	res, err := dao.AccessorFor(app.factory, client.NewGVR("contexts"))
	if err != nil {
		return nil
	}
	switcher, ok := res.(dao.Switchable)
	if !ok {
		return errors.New("Expecting a switchable resource")
	}
	if err := switcher.Switch(name); err != nil {
		log.Error().Err(err).Msgf("Context switch failed")
		return err
	}

	return app.switchCtx(name, true)
}
