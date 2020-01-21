package dialog

import (
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

const deleteKey = "delete"

type (
	okFunc     func(cascade, force bool)
	cancelFunc func()
)

// ShowDelete pops a resource deletion dialog.
func ShowDelete(pages *ui.Pages, msg string, ok okFunc, cancel cancelFunc) {
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
		dismissDelete(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		ok(cascade, force)
		dismissDelete(pages)
		cancel()
	})
	f.SetFocus(3)

	confirm := tview.NewModalForm("<Delete>", f)
	confirm.SetText(msg)
	confirm.SetDoneFunc(func(int, string) {
		dismissDelete(pages)
		cancel()
	})
	pages.AddPage(deleteKey, confirm, false, false)
	pages.ShowPage(deleteKey)
}

func dismissDelete(pages *ui.Pages) {
	pages.RemovePage(deleteKey)
}
