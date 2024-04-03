// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
)

const (
	renamePage = "rename"
	inputField = "New name:"
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
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *Context) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyR, ui.NewKeyAction("Rename", c.renameCmd, true))
}

func (c *Context) renameCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	c.showRenameModal(contextName, c.renameDialogCallback)

	return nil
}

func (c *Context) renameDialogCallback(form *tview.Form, contextName string) error {
	app := c.App()
	input := form.GetFormItemByLabel(inputField).(*tview.InputField)
	if err := app.factory.Client().Config().RenameContext(contextName, input.GetText()); err != nil {
		c.App().Flash().Err(err)
		return nil
	}
	c.Refresh()
	return nil
}

func (c *Context) showRenameModal(name string, ok func(form *tview.Form, contextName string) error) {
	app := c.App()
	styles := app.Styles.Dialog()

	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddInputField(inputField, name, 0, nil, nil).
		AddButton("OK", func() {
			if err := ok(f, name); err != nil {
				app.Flash().Err(err)
				return
			}
			app.Content.Pages.RemovePage(renamePage)
		}).
		AddButton("Cancel", func() {
			app.Content.Pages.RemovePage(renamePage)
		})

	m := tview.NewModalForm("<Rename>", f)
	m.SetText(fmt.Sprintf("Rename context %q?", name))
	m.SetDoneFunc(func(int, string) {
		app.Content.Pages.RemovePage(renamePage)
	})
	app.Content.Pages.AddPage(renamePage, m, false, false)
	app.Content.Pages.ShowPage(renamePage)

	for i := 0; i < f.GetButtonCount(); i++ {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
}

func (c *Context) useCtx(app *App, model ui.Tabular, gvr client.GVR, path string) {
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
		return errors.New("expecting a switchable resource")
	}
	if err := switcher.Switch(name); err != nil {
		log.Error().Err(err).Msgf("Context switch failed")
		return err
	}

	return app.switchContext(cmd.NewInterpreter("ctx "+name), true)
}
