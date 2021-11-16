package view

import (
	"strconv"
	"time"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

const drainKey = "drain"

// DrainFunc represents a drain callback function.
type DrainFunc func(v ResourceViewer, path string, opts dao.DrainOptions)

// ShowDrain pops a node drain dialog.
func ShowDrain(view ResourceViewer, path string, defaults dao.DrainOptions, okFn DrainFunc) {
	styles := view.App().Styles

	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.BgColor()).
		SetButtonTextColor(styles.FgColor()).
		SetLabelColor(styles.K9s.Info.FgColor.Color()).
		SetFieldTextColor(styles.K9s.Info.SectionColor.Color())

	var opts dao.DrainOptions
	f.AddInputField("GracePeriod:", strconv.Itoa(defaults.GracePeriodSeconds), 0, nil, func(v string) {
		a, err := asIntOpt(v)
		if err != nil {
			view.App().Flash().Err(err)
			return
		}
		view.App().Flash().Clear()
		opts.GracePeriodSeconds = a
	})
	f.AddInputField("Timeout:", defaults.Timeout.String(), 0, nil, func(v string) {
		a, err := asDurOpt(v)
		if err != nil {
			view.App().Flash().Err(err)
			return
		}
		view.App().Flash().Clear()
		opts.Timeout = a
	})
	f.AddCheckbox("Ignore DaemonSets:", defaults.IgnoreAllDaemonSets, func(_ string, v bool) {
		opts.IgnoreAllDaemonSets = v
	})
	f.AddCheckbox("Delete Local Data:", defaults.DeleteEmptyDirData, func(_ string, v bool) {
		opts.DeleteEmptyDirData = v
	})
	f.AddCheckbox("Force:", defaults.Force, func(_ string, v bool) {
		opts.Force = v
	})

	pages := view.App().Content.Pages
	f.AddButton("Cancel", func() {
		DismissDrain(view, pages)
	})
	f.AddButton("OK", func() {
		DismissDrain(view, pages)
		okFn(view, path, opts)
	})

	modal := tview.NewModalForm("<Drain>", f)
	modal.SetText(path)
	modal.SetDoneFunc(func(_ int, b string) {
		DismissDrain(view, pages)
	})

	pages.AddPage(drainKey, modal, false, true)
	pages.ShowPage(drainKey)
	view.App().SetFocus(pages.GetPrimitive(drainKey))
}

// DismissDrain dismiss the port forward dialog.
func DismissDrain(v ResourceViewer, p *ui.Pages) {
	p.RemovePage(drainKey)
	v.App().SetFocus(p.CurrentPage().Item)
}

// ----------------------------------------------------------------------------
// Helpers...

func asDurOpt(v string) (time.Duration, error) {
	d, err := time.ParseDuration(v)
	if err != nil {
		return 0, err
	}

	return d, nil
}

func asIntOpt(v string) (int, error) {
	i, err := strconv.Atoi(v)
	if err != nil {
		return 0, err
	}

	return i, nil
}
