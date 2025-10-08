// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

type TextFn func(value *string)

type TextDialogOpts struct {
	Title           string
	Text            string
	Label           string
	Placeholder     string
	InitialValue    string
	MaxSize         int
	AllowEmptyValue bool
	Selected        TextFn
	Cancel          cancelFunc
}

func ShowText(styles *config.Dialog, pages *ui.Pages, opts *TextDialogOpts) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	value := opts.InitialValue
	f.AddInputField(opts.Label, opts.InitialValue, opts.MaxSize, nil, nil)
	valueField := f.GetFormItemByLabel(opts.Label).(*tview.InputField)
	if valueField.GetText() == "" {
		valueField.SetPlaceholder(opts.Placeholder)
	}
	valueField.SetChangedFunc(func(s string) {
		value = s
		if value == "" {
			valueField.SetPlaceholder(opts.Placeholder)
		} else {
			valueField.SetPlaceholder("")
		}
	})

	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		opts.Cancel()
	})
	f.AddButton("OK", func() {
		if value == "" && !opts.AllowEmptyValue {
			return
		}
		opts.Selected(&value)
		dismissConfirm(pages)
	})

	for i := range 2 {
		b := f.GetButton(i)
		if b == nil {
			continue
		}
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
	f.SetFocus(0)

	modal := tview.NewModalForm("<"+opts.Title+">", f)
	modal.SetText(opts.Text)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetBackgroundColor(styles.BgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		opts.Cancel()
	})

	pages.AddPage(confirmKey, modal, false, true)
	pages.ShowPage(confirmKey)
}
