package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const logKey = "logs"

// LogCB represents a log callback function.
type LogCB func(path string, opts dao.LogOptions)

// ShowLogs pops a port forwarding configuration dialog.
func ShowLogs(a *App, path string, applyFn LogCB) {
	styles := a.Styles

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.BgColor()).
		SetButtonTextColor(styles.FgColor()).
		SetLabelColor(styles.K9s.Info.FgColor.Color()).
		SetFieldTextColor(styles.K9s.Info.SectionColor.Color())

	secs, start, in, out, container := "5", time.Now().String(), "", "", ""
	f.AddInputField("Container:", container, 0, nil, func(v string) {
		container = v
	})
	f.AddInputField("Since Seconds:", secs, 0, nil, func(v string) {
		secs = v
	})
	f.AddInputField("Since Time:", start, 0, nil, func(v string) {
		start = v
	})
	f.AddInputField("Filter In:", in, 0, nil, func(v string) {
		in = v
	})
	f.AddInputField("Filter Out:", out, 0, nil, func(v string) {
		out = v
	})

	pages := a.Content.Pages

	f.AddButton("Apply", func() {
		s, _ := strconv.Atoi(secs)
		opts := dao.LogOptions{
			SinceTime:    start,
			SinceSeconds: int64(s),
			In:           in,
			Out:          out,
		}
		applyFn(path, opts)
	})
	f.AddButton("Dismiss", func() {
		DismissLogs(a, pages)
	})

	modal := tview.NewModalForm(fmt.Sprintf("<Configure Logs for %s>", path), f)
	modal.SetDoneFunc(func(_ int, b string) {
		DismissLogs(a, pages)
	})

	pages.AddPage(logKey, modal, false, true)
	pages.ShowPage(logKey)
	a.SetFocus(pages.GetPrimitive(logKey))
}

// DismissLogs dismiss the dialog.
func DismissLogs(a *App, p *ui.Pages) {
	p.RemovePage(logKey)
	a.SetFocus(p.CurrentPage().Item)
}

// ----------------------------------------------------------------------------
// Helpers...
