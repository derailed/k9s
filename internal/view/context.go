package view

import (
	"errors"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

const renamePage = "rename"

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
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *Context) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyR: ui.NewKeyAction("Rename", c.renameCmd, true),
	})
}

func (c *Context) renameCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	app := c.App()
	c.showRenameModal(app, contextName, func(newName string)(error) {
		if err := app.factory.Client().Config().RenameContext(contextName, newName); err != nil {
			c.App().Flash().Err(err)
			return nil
		}
		c.Refresh()
		return nil
	})

	return nil
}

func (c *Context) showRenameModal(a *App, msg string, ok func(newName string)(error)) {
	p := a.Content.Pages
	f := c.makeStyledForm()
	f.AddInputField("New name:", msg, 0, nil, nil)
	contextField := f.GetFormItemByLabel("New name:").(*tview.InputField)
	f.AddButton("OK", func() {
		if err := ok(contextField.GetText()); err != nil {
			c.App().Flash().Err(err)
			return
		}
		p.RemovePage(renamePage)
	})
	f.AddButton("Cancel", func() {
		p.RemovePage(renamePage)
	})
	m := tview.NewModalForm("<Rename>", f)
	p.AddPage(renamePage, m, false, false)
	p.ShowPage(renamePage)
}

func (c *Context) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	return f
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
		return err
	}
	switcher, ok := res.(dao.Switchable)
	if !ok {
		return errors.New("Expecting a switchable resource")
	}
	if err := switcher.Switch(name); err != nil {
		log.Error().Err(err).Msgf("Context switch failed")
		return err
	}

	return app.switchContext(name, true)
}
