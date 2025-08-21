// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"
	"fmt"
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/k9s/internal/view/cmd"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
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
func NewContext(gvr *client.GVR) ResourceViewer {
	c := Context{
		ResourceViewer: NewBrowser(gvr),
	}
	c.GetTable().SetEnterFn(c.useCtx)
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *Context) bindKeys(aa *ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, tcell.KeyCtrlSpace, ui.KeySpace)
	if !c.App().Config.IsReadOnly() {
		c.bindDangerousKeys(aa)
	}
}

func (c *Context) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyR, ui.NewKeyAction("Rename", c.renameCmd, true))
	aa.Add(tcell.KeyCtrlD, ui.NewKeyAction("Delete", c.deleteCmd, true))
}

func (c *Context) renameCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	c.showRenameModal(contextName, c.renameDialogCallback)

	return nil
}

func (c *Context) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	contextName := c.GetTable().GetSelectedItem()
	if contextName == "" {
		return evt
	}

	d := c.App().Styles.Dialog()
	dialog.ShowConfirm(&d, c.App().Content.Pages, "Delete", fmt.Sprintf("Delete context %q?", contextName), func() {
		if err := c.App().factory.Client().Config().DelContext(contextName); err != nil {
			c.App().Flash().Err(err)
			return
		}
		c.Refresh()
	}, func() {})

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
			app.Content.RemovePage(renamePage)
		})

	m := tview.NewModalForm("<Rename>", f)
	m.SetText(fmt.Sprintf("Rename context %q?", name))
	m.SetDoneFunc(func(int, string) {
		app.Content.RemovePage(renamePage)
	})
	app.Content.AddPage(renamePage, m, false, false)
	app.Content.ShowPage(renamePage)

	for i := range f.GetButtonCount() {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
}

func (c *Context) useCtx(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	slog.Debug("Using context",
		slogs.GVR, gvr,
		slogs.FQN, path,
	)
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
	res, err := dao.AccessorFor(app.factory, client.CtGVR)
	if err != nil {
		return err
	}
	switcher, ok := res.(dao.Switchable)
	if !ok {
		return errors.New("expecting a switchable resource")
	}

	app.Config.K9s.ToggleContextSwitch(true)
	defer app.Config.K9s.ToggleContextSwitch(false)

	// Save config prior to context switch...
	if err := app.Config.Save(true); err != nil {
		slog.Error("Fail to save config to disk", slogs.Subsys, "config", slogs.Error, err)
	}

	if err := switcher.Switch(name); err != nil {
		slog.Error("Context switch failed during use command", slogs.Error, err)
		return err
	}

	return app.switchContext(cmd.NewInterpreter("ctx "+name), true)
}
