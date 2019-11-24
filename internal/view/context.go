package view

import (
	"context"
	"errors"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// Context presents a context viewer.
type Context struct {
	ResourceViewer
}

// NewContext return a new context viewer.
func NewContext(title, gvr string, list resource.List) ResourceViewer {
	return &Context{
		ResourceViewer: NewResource(title, gvr, list).(ResourceViewer),
	}
}

func (c *Context) Init(ctx context.Context) error {
	c.GetTable().SetEnterFn(c.useCtx)
	if err := c.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	c.GetTable().SetSelectedFn(c.cleanser)
	c.bindKeys()

	return nil
}

func (c *Context) bindKeys() {
	c.Actions().Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
}

func (c *Context) useCtx(app *App, _, res, sel string) {
	if err := c.useContext(sel); err != nil {
		app.Flash().Err(err)
		return
	}
	if !app.gotoResource("po") {
		app.Flash().Err(errors.New("Goto pod failed"))
	}
}

func (*Context) cleanser(s string) string {
	name := strings.TrimSpace(s)
	if strings.HasSuffix(name, "*") {
		name = strings.TrimRight(name, "*")
	}
	if strings.HasSuffix(name, "(𝜟)") {
		name = strings.TrimRight(name, "(𝜟)")
	}
	return name
}

func (c *Context) useContext(name string) error {
	ctx := c.cleanser(name)
	if err := c.List().Resource().(*resource.Context).Switch(ctx); err != nil {
		return err
	}

	if err := c.App().switchCtx(name, false); err != nil {
		return err
	}
	c.Refresh()
	c.GetTable().Select(1, 0)

	return nil
}
