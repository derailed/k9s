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

func (c *Context) useCtx(app *App, _, res, path string) {
	log.Debug().Msgf("SWITCH CTX %q--%q", res, path)
	if err := c.useContext(path); err != nil {
		app.Flash().Err(err)
		return
	}
	if err := app.gotoResource("po", true); err != nil {
		app.Flash().Err(err)
	}
}

func (c *Context) useContext(name string) error {
	res, err := dao.AccessorFor(c.App().factory, client.NewGVR(c.GVR()))
	if err != nil {
		return nil
	}

	switcher, ok := res.(dao.Switchable)
	if !ok {
		return errors.New("Expecting a switchable resource")
	}
	if err := switcher.Switch(name); err != nil {
		return err
	}
	if err := c.App().switchCtx(name, false); err != nil {
		return err
	}
	c.Refresh()
	c.GetTable().Select(1, 0)

	return nil
}
