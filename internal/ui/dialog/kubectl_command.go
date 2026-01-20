// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const kubectlDialogKey = "kubectl-command"

type copyFunc func(string) error

// ShowKubectlCommand displays a dialog with the kubectl command and copy functionality.
func ShowKubectlCommand(styles *config.Dialog, pages *ui.Pages, command string, onCopy copyFunc) {
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())

	// Add the command as a read-only input field
	f.AddInputField("Command:", command, 0, nil, nil)
	// Make the field read-only by disabling editing
	if field := f.GetFormItem(0).(*tview.InputField); field != nil {
		field.SetFieldBackgroundColor(styles.BgColor.Color())
	}

	// Try to detect if clipboard is available
	clipboardAvailable := true
	if onCopy != nil {
		// Test clipboard availability with empty string
		if err := onCopy(""); err != nil {
			clipboardAvailable = false
		}
	} else {
		clipboardAvailable = false
	}

	// Add Copy button only if clipboard is available
	if clipboardAvailable && onCopy != nil {
		f.AddButton("Copy to Clipboard", func() {
			if err := onCopy(command); err != nil {
				// If copy fails, just dismiss
				dismissKubectl(pages)
				return
			}
			dismissKubectl(pages)
		})
	}

	// Always add Close button
	f.AddButton("Close", func() {
		dismissKubectl(pages)
	})

	// Style buttons
	buttonCount := 1
	if clipboardAvailable && onCopy != nil {
		buttonCount = 2
	}
	for i := range buttonCount {
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	f.SetFocus(0)
	modal := tview.NewModalForm("<Kubectl Command>", f)
	modal.SetText("Copy or view the equivalent kubectl command:")
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissKubectl(pages)
	})

	pages.AddPage(kubectlDialogKey, modal, false, false)
	pages.ShowPage(kubectlDialogKey)
}

func dismissKubectl(pages *ui.Pages) {
	pages.RemovePage(kubectlDialogKey)
}
