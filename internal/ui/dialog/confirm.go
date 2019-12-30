package dialog

import (
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const confirmKey = "confirm"

type (
	confirmFunc func()
)

// ShowConfirm pops a confirmation dialog.
func ShowConfirm(pages *ui.Pages, title, msg string, ack confirmFunc, cancel cancelFunc) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)
	f.AddButton("Cancel", func() {
		dismissConfirm(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		ack()
		dismissConfirm(pages)
		cancel()
	})

	modal := tview.NewModalForm(" <"+title+"> ", f)
	modal.SetText(msg)
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
