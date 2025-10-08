// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

type selectedFunc func(index int, value string)

type ListSelectionDialogOpts struct {
	Title                string
	Text                 string
	Label                string
	Options              []string
	InitialValueIndex    int
	AllowUnSelectedValue bool
	Selected             selectedFunc
	Cancel               cancelFunc
}

func ShowListSelection(styles *config.Dialog, pages *ui.Pages, opts *ListSelectionDialogOpts) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	selectedIndex := opts.InitialValueIndex
	// We confirm that initial index is not out of array size
	if selectedIndex >= len(opts.Options) {
		selectedIndex = -1
	}

	selectedValue := ""
	// We set the selected value to the index if it's not out of array size
	if selectedIndex >= 0 {
		selectedValue = opts.Options[selectedIndex]
	}

	f.AddDropDown(opts.Label, opts.Options, selectedIndex, func(_ string, optionIndex int) {
		selectedValue = opts.Options[optionIndex]
		selectedIndex = optionIndex
	})
	selectionField := f.GetFormItemByLabel(opts.Label).(*tview.DropDown)
	selectionField.SetListStyles(
		styles.FgColor.Color(), styles.BgColor.Color(),
		styles.ButtonFocusFgColor.Color(), styles.ButtonFocusBgColor.Color(),
	)

	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		opts.Cancel()
	})
	f.AddButton("OK", func() {
		if selectedIndex < 0 && !opts.AllowUnSelectedValue {
			return
		}
		opts.Selected(selectedIndex, selectedValue)
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
