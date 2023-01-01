package dialog

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const deleteKey = "delete"

const noDeletePropagation = "None"

const defaultPropagationIdx = 0

var propagationOptions []string = []string{
	string(metav1.DeletePropagationBackground),
	string(metav1.DeletePropagationForeground),
	string(metav1.DeletePropagationOrphan),
	noDeletePropagation,
}

type (
	okFunc     func(propagation *metav1.DeletionPropagation, force bool)
	cancelFunc func()
)

// ShowDelete pops a resource deletion dialog.
func ShowDelete(styles config.Dialog, pages *ui.Pages, msg string, ok okFunc, cancel cancelFunc) {
	propagation, force := "", false
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddDropDown("Propagation:", propagationOptions, defaultPropagationIdx, func(_ string, optionIndex int) {
		propagation = propagationOptions[optionIndex]
	})
	propField := f.GetFormItemByLabel("Propagation:").(*tview.DropDown)
	propField.SetListStyles(
		styles.FgColor.Color(),
		styles.BgColor.Color(),
		styles.ButtonFocusFgColor.Color(),
		styles.ButtonFocusBgColor.Color(),
	)
	f.AddCheckbox("Force:", force, func(_ string, checked bool) {
		force = checked
	})
	f.AddButton("Cancel", func() {
		dismissDelete(pages)
		cancel()
	})
	f.AddButton("OK", func() {
		switch propagation {
		case noDeletePropagation:
			ok(nil, force)
		default:
			p := metav1.DeletionPropagation(propagation)
			ok(&p, force)
		}
		dismissDelete(pages)
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
	f.SetFocus(2)

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
