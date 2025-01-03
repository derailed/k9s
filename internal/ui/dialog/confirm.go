// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const dialogKey = "dialog"

type confirmFunc func()

func ShowConfirmAck(app *ui.App, pages *ui.Pages, acceptStr string, override bool, title, msg string, ack confirmFunc, cancel cancelFunc) {
	styles := app.Styles.Dialog()

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		cancel()
	})

	var accept bool
	if override {
		changedFn := func(t string) {
			accept = (t == acceptStr)
		}
		f.AddInputField("Confirm:", "", 30, nil, changedFn)
	} else {
		accept = true
	}

	f.AddButton("OK", func() {
		if !accept {
			return
		}
		ack()
		dismissConfirm(pages)
		cancel()
	})
	for i := 0; i < 2; i++ {
		b := f.GetButton(i)
		if b == nil {
			continue
		}
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
	f.SetFocus(0)
	modal := tview.NewModalForm("<"+title+">", f)
	modal.SetText(msg)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		cancel()
	})
	pages.AddPage(confirmKey, modal, false, false)
	pages.ShowPage(confirmKey)
}

// ShowConfirm pops a confirmation dialog.
func ShowConfirm(styles config.Dialog, pages *ui.Pages, title, msg string, ack confirmFunc, cancel cancelFunc) {
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())
	f.AddButton("Cancel", func() {
		dismiss(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		ack()
		dismiss(pages)
		cancel()
	})
	for i := 0; i < 2; i++ {
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}
	f.SetFocus(0)
	modal := tview.NewModalForm("<"+title+">", f)
	modal.SetText(msg)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
		cancel()
	})
	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)
}

func dismiss(pages *ui.Pages) {
	pages.RemovePage(dialogKey)
}
