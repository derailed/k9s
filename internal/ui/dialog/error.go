package dialog

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

// ShowError pops an error dialog.
func ShowError(styles config.Dialog, pages *ui.Pages, msg string) {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())
	f.AddButton("Dismiss", func() {
		dismiss(pages)
	})
	if b := f.GetButton(0); b != nil {
		b.SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color())
		b.SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}
	f.SetFocus(0)
	modal := tview.NewModalForm("<error>", f)
	modal.SetText(cowTalk(msg))
	modal.SetTextColor(styles.ErrorColor.Color())
	modal.SetDoneFunc(func(int, string) {
		dismiss(pages)
	})
	pages.AddPage(dialogKey, modal, false, false)
	pages.ShowPage(dialogKey)
}

func cowTalk(says string) string {
	msg := fmt.Sprintf("< Ruroh? %s >", says)
	buff := make([]string, 0, len(cow)+3)
	buff = append(buff, msg)
	buff = append(buff, cow...)

	return strings.Join(buff, "\n")
}

var cow = []string{
	`\   ^__^            `,
	` \  (oo)\_______    `,
	`    (__)\       )\/\`,
	`        ||----w |   `,
	`        ||     ||   `,
}
