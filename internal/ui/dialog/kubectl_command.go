// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const kubectlDialogKey = "kubectl-command"

type copyFunc func(string) error

// ShowKubectlCommand displays a dialog with the kubectl command and copy functionality.
func ShowKubectlCommand(styles *config.Dialog, pages *ui.Pages, command string, onCopy copyFunc, flash *model.Flash) {
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

	// Always add Copy button (best-effort approach like other dialogs)
	f.AddButton("Copy to Clipboard", func() {
		if err := onCopy(command); err != nil {
			flash.Err(err)
			return
		}
		flash.Info("Command copied to clipboard...")
		dismissKubectl(pages)
	})

	// Always add Close button
	f.AddButton("Close", func() {
		dismissKubectl(pages)
	})

	// Style buttons
	for i := range 2 {
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
