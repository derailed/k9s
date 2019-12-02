package view

import (
	"errors"
	"strings"

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

// NewContext return a new context viewer.
func NewContext(gvr dao.GVR) ResourceViewer {
	c := Context{
		ResourceViewer: NewGeneric(gvr),
	}
	c.GetTable().SetEnterFn(c.useCtx)
	c.GetTable().SetSelectedFn(c.cleanser)
	c.GetTable().SetColorerFn(render.Context{}.ColorerFunc())
	c.BindKeys()

	return &c
}

func (c *Context) BindKeys() {
	c.Actions().Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
}

func (c *Context) useCtx(app *App, _, res, sel string) {
	if err := c.useContext(sel); err != nil {
		app.Flash().Err(err)
		return
	}
	if !app.gotoResource("po") {
		app.Flash().Err(errors.New("goto pod failed"))
	}
}

func (*Context) cleanser(s string) string {
	name := strings.TrimSpace(s)
	if strings.HasSuffix(name, "(*)") {
		name = strings.TrimRight(name, "(*)")
	}
	if strings.HasSuffix(name, "(𝜟)") {
		name = strings.TrimRight(name, "(𝜟)")
	}
	return name
}

func (c *Context) useContext(name string) error {
	res, err := dao.AccessorFor(c.App().factory, dao.GVR(c.GVR()))
	if err != nil {
		return nil
	}

	switcher, ok := res.(dao.Switchable)
	if !ok {
		return errors.New("Expecting a switchable resource")
	}

	log.Debug().Msgf("Context %q", name)
	ctx, _ := namespaced(name)
	ctx = c.cleanser(ctx)
	if err := switcher.Switch(ctx); err != nil {
		return err
	}

	if err := c.App().switchCtx(ctx, false); err != nil {
		return err
	}
	c.Refresh()
	c.GetTable().Select(1, 0)

	return nil
}
