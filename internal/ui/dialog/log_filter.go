// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

type shellFilterFunc func(cmd string)

// ShowLogFilter pops a dialog that lets the user enter (or clear) a shell
// filter command for the log stream.  The current filter value is shown as the
// pre-filled field text.  Submitting an empty string clears the filter.
func ShowLogFilter(styles *config.Dialog, pages *ui.Pages, current string, ack shellFilterFunc, cancel cancelFunc) {
	cmd := current

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	f.AddInputField("Filter:", current, 50, nil, func(v string) {
		cmd = v
	})

	f.AddButton("Cancel", func() {
		dismiss(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		ack(cmd)
		dismiss(pages)
		cancel()
	})
	for i := range f.GetButtonCount() {
		b := f.GetButton(i)
		if b == nil {
			continue
		}
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
	f.SetFocus(0)

	modal := tview.NewModalForm("<Shell Filter>", f)
	modal.SetText("Pipe log lines through a shell command.\nLeave empty to disable.")
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
		cancel()
	})
	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)
}
