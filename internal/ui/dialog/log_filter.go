// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

type shellFilterFunc func(cmd string)

// ShowLogFilter pops a dialog that lets the user enter (or clear) a shell
// filter command for the log stream.  Up/Down arrows cycle through history.
// Submitting an empty string clears the filter.
func ShowLogFilter(styles *config.Dialog, pages *ui.Pages, current string, history []string, ack shellFilterFunc, cancel cancelFunc) {
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

	// Wire Up/Down arrows on the input field to cycle through history.
	// navIdx == len(history) means the user is at their current typed value.
	if field, ok := f.GetFormItemByLabel("Filter:").(*tview.InputField); ok {
		navIdx := len(history)
		savedCurrent := current
		field.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
			switch event.Key() {
			case tcell.KeyUp:
				if navIdx > 0 {
					if navIdx == len(history) {
						savedCurrent = field.GetText()
					}
					navIdx--
					text := history[len(history)-1-navIdx]
					field.SetText(text)
					cmd = text
				}
				return nil
			case tcell.KeyDown:
				if navIdx < len(history) {
					navIdx++
					var text string
					if navIdx == len(history) {
						text = savedCurrent
					} else {
						text = history[len(history)-1-navIdx]
					}
					field.SetText(text)
					cmd = text
				}
				return nil
			}
			return event
		})
	}

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
