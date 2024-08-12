// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const confirmKey = "confirm"

type TransferFn func(TransferArgs) bool

type TransferArgs struct {
	From, To, CO         string
	Download, NoPreserve bool
	Retries              int
}

type TransferDialogOpts struct {
	Containers     []string
	Pod            string
	Title, Message string
	Retries        int
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

	args := TransferArgs{
		From:    opts.Pod,
		Retries: opts.Retries,
	}
	var fromField, toField *tview.InputField
	args.Download = true
	f.AddCheckbox("Download:", args.Download, func(_ string, flag bool) {
		if flag {
			modal.SetText(strings.Replace(opts.Message, "Upload", "Download", 1))
		} else {
			modal.SetText(strings.Replace(opts.Message, "Download", "Upload", 1))
		}
		args.Download = flag
		args.From, args.To = args.To, args.From
		fromField.SetText(args.From)
		toField.SetText(args.To)
	})

	f.AddInputField("From:", args.From, 40, nil, func(v string) {
		args.From = v
	})
	f.AddInputField("To:", args.To, 40, nil, func(v string) {
		args.To = v
	})
	fromField, _ = f.GetFormItemByLabel("From:").(*tview.InputField)
	toField, _ = f.GetFormItemByLabel("To:").(*tview.InputField)

	f.AddCheckbox("NoPreserve:", args.NoPreserve, func(_ string, f bool) {
		args.NoPreserve = f
	})
	if len(opts.Containers) > 0 {
		args.CO = opts.Containers[0]
	}
	f.AddInputField("Container:", args.CO, 30, nil, func(v string) {
		args.CO = v
	})
	retries := strconv.Itoa(opts.Retries)
	f.AddInputField("Retries:", retries, 30, nil, func(v string) {
		retries = v

		if retriesInt, err := strconv.Atoi(retries); err == nil {
			args.Retries = retriesInt
		}
	})

	f.AddButton("OK", func() {
		if !opts.Ack(args) {
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
