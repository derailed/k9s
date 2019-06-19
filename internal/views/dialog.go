package views

import (
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const deleteDialogKey = "delete"

type (
	doneFn   func(cascade, force bool)
	cancelFn func()
)

func showDeleteDialog(pages *tview.Pages, msg string, done doneFn, cancel cancelFn) {
	cascade, force := true, false
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)
	f.AddCheckbox("Cascade:", cascade, func(checked bool) {
		cascade = checked
	})
	f.AddCheckbox("Force:", force, func(checked bool) {
		force = checked
	})
	f.AddButton("Cancel", func() {
		dismissDeleteDialog(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		done(cascade, force)
		dismissDeleteDialog(pages)
		cancel()
	})

	confirm := tview.NewModalForm("<Delete>", f)
	confirm.SetText(msg)
	confirm.SetDoneFunc(func(int, string) {
		dismissDeleteDialog(pages)
		cancel()
	})
	pages.AddPage(deleteDialogKey, confirm, false, false)
	pages.ShowPage(deleteDialogKey)
}

func dismissDeleteDialog(pages *tview.Pages) {
	pages.RemovePage(deleteDialogKey)
}
