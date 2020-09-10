package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const confirmKey = "confirm"

type (
	confirmFunc func()
)

// ShowConfirm pops a confirmation dialog.
func ShowConfirm(styles config.Dialog, pages *ui.Pages, title, msg string, ack confirmFunc, cancel cancelFunc) {
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
	f.AddButton("OK", func() {
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
	modal.SetBackgroundColor(styles.BgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismissConfirm(pages)
		cancel()
	})
	pages.AddPage(confirmKey, modal, false, false)
	pages.ShowPage(confirmKey)
}

func dismissConfirm(pages *ui.Pages) {
	pages.RemovePage(confirmKey)
}
