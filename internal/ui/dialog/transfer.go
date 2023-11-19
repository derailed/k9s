// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const confirmKey = "confirm"

type TransferFn func(from, to, co string, download, no_preserve bool) bool

type TransferDialogOpts struct {
	Containers     []string
	Pod            string
	Title, Message string
	Ack            TransferFn
	Cancel         cancelFunc
}

func ShowUploads(styles config.Dialog, pages *ui.Pages, opts TransferDialogOpts) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		opts.Cancel()
	})

	modal := tview.NewModalForm("<"+opts.Title+">", f)

	from, to := opts.Pod, ""
	var fromField, toField *tview.InputField
	download := true
	f.AddCheckbox("Download:", download, func(_ string, flag bool) {
		if flag {
			modal.SetText(strings.Replace(opts.Message, "Upload", "Download", 1))
		} else {
			modal.SetText(strings.Replace(opts.Message, "Download", "Upload", 1))
		}
		download = flag
		from, to = to, from
		fromField.SetText(from)
		toField.SetText(to)
	})

	f.AddInputField("From:", from, 40, nil, func(t string) {
		from = t
	})
	f.AddInputField("To:", to, 40, nil, func(t string) {
		to = t
	})
	fromField, _ = f.GetFormItemByLabel("From:").(*tview.InputField)
	toField, _ = f.GetFormItemByLabel("To:").(*tview.InputField)

	var no_preserve bool
	f.AddCheckbox("NoPreserve:", no_preserve, func(_ string, f bool) {
		no_preserve = f
	})
	var co string
	if len(opts.Containers) > 0 {
		co = opts.Containers[0]
	}
	f.AddInputField("Container:", co, 30, nil, func(t string) {
		co = t
	})

	f.AddButton("OK", func() {
		if !opts.Ack(from, to, co, download, no_preserve) {
			return
		}
		dismissConfirm(pages)
		opts.Cancel()
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

	message := opts.Message
	if len(opts.Containers) > 1 {
		message += "\nAvailable Containers:" + strings.Join(opts.Containers, ",")
	}
	modal.SetText(message)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		opts.Cancel()
	})
	pages.AddPage(confirmKey, modal, false, false)
	pages.ShowPage(confirmKey)
}

func dismissConfirm(pages *ui.Pages) {
	pages.RemovePage(confirmKey)
}
