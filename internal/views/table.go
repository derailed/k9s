package views

import (
	"github.com/derailed/k9s/internal/filters"

	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type tableView struct {
	*ui.Table

	app           *appView
	labelFilterFn func(string)
}

func newTableView(app *appView, title string) *tableView {
	v := tableView{
		Table: ui.NewTable(title, app.Styles, &app.stickyFilter),
		app:   app,
	}
	app.Cmd().Clear()
	v.SearchBuff().AddListener(app.Cmd(), &v)
	v.bindKeys()

	if v.app.stickyFilter {
		v.SearchBuff().Set(v.app.filter)
	}
	return &v
}

// BufferChanged indicates the buffer was changed.
func (v *tableView) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (v *tableView) BufferActive(state bool, k ui.BufferKind) {
	v.app.BufferActive(state, k)
}

func (v *tableView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveTable(v.app.Config.K9s.CurrentCluster, v.GetBaseTitle(), v.GetFilteredData()); err != nil {
		v.app.Flash().Err(err)
	} else {
		v.app.Flash().Infof("File %s saved successfully!", path)
	}

	return nil
}

func (v *tableView) setLabelFilterFn(fn func(string)) {
	v.labelFilterFn = fn
	if v.labelFilterFn == nil {
		return
	}

	cmd := v.SearchBuff().String()
	if q, ok := filters.LabelSelector(cmd); ok {
		v.labelFilterFn(q)
	}
}

func (v *tableView) bindKeys() {
	v.SetActions(ui.KeyActions{
		tcell.KeyCtrlS:      ui.NewKeyAction("Save", v.saveCmd, true),
		ui.KeySlash:         ui.NewKeyAction("Filter Mode", v.activateCmd, false),
		tcell.KeyEscape:     ui.NewKeyAction("Filter Reset", v.resetCmd, false),
		tcell.KeyEnter:      ui.NewKeyAction("Filter", v.filterCmd, false),
		tcell.KeyBackspace2: ui.NewKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyBackspace:  ui.NewKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyDelete:     ui.NewKeyAction("Erase", v.eraseCmd, false),
		ui.KeyShiftI:        ui.NewKeyAction("Invert", v.SortInvertCmd, false),
		ui.KeyShiftN:        ui.NewKeyAction("Sort Name", v.SortColCmd(0), false),
		ui.KeyShiftA:        ui.NewKeyAction("Sort Age", v.SortColCmd(-1), false),
	})
}

func (v *tableView) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.SearchBuff().IsActive() {
		return evt
	}

	v.SearchBuff().SetActive(false)

	filter := v.SearchBuff().String()
	v.app.updateFilter(filter)

	if v.labelFilterFn == nil {
		v.Refresh()
		return nil
	}

	if q, ok := filters.LabelSelector(filter); ok {
		v.labelFilterFn(q)
	} else {
		v.labelFilterFn("")
	}

	v.Refresh()
	return nil
}

func (v *tableView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.SearchBuff().IsActive() {
		v.SearchBuff().Delete()
	}

	return nil
}

func (v *tableView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.SearchBuff().Empty() {
		v.app.Flash().Info("Clearing filter...")
	}

	if !v.app.stickyFilter {
		v.app.updateFilter("")
	}

	if _, ok := filters.LabelSelector(v.SearchBuff().String()); ok {
		v.labelFilterFn("")
	}
	v.SearchBuff().Reset()
	v.Refresh()

	return nil
}

func (v *tableView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.InCmdMode() {
		return evt
	}

	v.app.Flash().Info("Filter mode activated.")
	v.SearchBuff().SetActive(true)
	return nil
}
