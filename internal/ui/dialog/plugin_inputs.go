// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const pluginInputsKey = "pluginInputs"

// PluginInputValues holds the collected input values from the dialog.
type PluginInputValues map[string]string

// PluginInputsOkFunc is called when the user confirms the plugin inputs.
type PluginInputsOkFunc func(values PluginInputValues)

// PluginInputsFlashFunc is called to display flash messages.
type PluginInputsFlashFunc func(msg string)

// ShowPluginInputs pops a dialog to collect plugin input values.
func ShowPluginInputs(
	styles *config.Dialog,
	pages *ui.Pages,
	title string,
	inputs []config.PluginInput,
	flash PluginInputsFlashFunc,
	ok PluginInputsOkFunc,
	cancel cancelFunc,
) {
	if len(inputs) == 0 {
		ok(make(PluginInputValues))
		return
	}

	values := make(PluginInputValues)

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	// Add input fields based on type
	for _, input := range inputs {
		label := input.Name
		if input.Label != "" {
			label = input.Label
		}
		if input.Required {
			label += " *"
		}
		label += ":"

		switch input.Type {
		case config.InputTypeString:
			values[input.Name] = ""
			inputName := input.Name
			f.AddInputField(label, "", 40, nil, func(text string) {
				values[inputName] = text
			})

		case config.InputTypeNumber:
			values[input.Name] = ""
			inputName := input.Name
			f.AddInputField(label, "", 20, func(text string, _ rune) bool {
				// Allow empty, negative sign, dot for decimals, or valid numbers
				if text == "" || text == "-" || text == "." || text == "-." {
					return true
				}
				_, err := strconv.ParseFloat(text, 64)
				return err == nil
			}, func(text string) {
				values[inputName] = text
			})

		case config.InputTypeBool:
			values[input.Name] = "false"
			inputName := input.Name
			f.AddCheckbox(label, false, func(_ string, checked bool) {
				values[inputName] = fmt.Sprintf("%t", checked)
			})

		case config.InputTypeDropdown:
			if len(input.Options) > 0 {
				values[input.Name] = ""
				inputName := input.Name
				// Prepend empty option so dropdown starts unselected
				options := append([]string{""}, input.Options...)
				f.AddDropDown(label, options, 0, func(_ string, optionIndex int) {
					if optionIndex >= 0 && optionIndex < len(options) {
						values[inputName] = options[optionIndex]
					}
				})

				if dropDown := f.GetFormItemByLabel(label); dropDown != nil {
					if dd, ok := dropDown.(*tview.DropDown); ok {
						dd.SetListStyles(
							styles.FgColor.Color(), styles.BgColor.Color(),
							styles.ButtonFocusFgColor.Color(), styles.ButtonFocusBgColor.Color(),
						)
					}
				}
			}
		}
	}

	// Add Cancel button
	f.AddButton("Cancel", func() {
		dismissPluginInputs(pages)
		cancel()
	})
	// Add OK button with validation
	f.AddButton("OK", func() {
		// Validate required fields
		var missing []string
		for _, input := range inputs {
			if input.Required {
				val := values[input.Name]
				// Bools always have a value (true/false), so skip validation for them
				if input.Type != config.InputTypeBool && val == "" {
					missing = append(missing, input.Name)
				}
			}
		}
		if len(missing) > 0 {
			if flash != nil {
				flash("Required fields are missing")
			}
			return
		}

		// Remove optional fields with zero values
		for _, input := range inputs {
			if !input.Required && input.Type != config.InputTypeBool && values[input.Name] == "" {
				delete(values, input.Name)
			}
		}

		ok(values)
		dismissPluginInputs(pages)
		cancel()
	})

	// Style buttons
	buttonCount := f.GetButtonCount()
	for i := range buttonCount {
		if b := f.GetButton(i); b != nil {
			b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
			b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
		}
	}

	f.SetFocus(0)

	modal := tview.NewModalForm("<"+title+">", f)
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissPluginInputs(pages)
		cancel()
	})

	pages.AddPage(pluginInputsKey, modal, false, false)
	pages.ShowPage(pluginInputsKey)
}

func dismissPluginInputs(pages *ui.Pages) {
	pages.RemovePage(pluginInputsKey)
}
