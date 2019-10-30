package views

import (
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type tableView struct {
	*ui.Table

	app      *appView
	filterFn func(string)
}

func newTableView(app *appView, title string) *tableView {
	v := tableView{
		Table: ui.NewTable(title, app.Styles),
		app:   app,
	}
	v.SearchBuff().AddListener(app.Cmd())
	v.SearchBuff().AddListener(&v)
	v.bindKeys()

	return &v
}

// BufferChanged indicates the buffer was changed.
func (v *tableView) BufferChanged(s string) {}

// BufferActive indicates the buff activity changed.
func (v *tableView) BufferActive(state bool, k ui.BufferKind) {
	v.app.BufferActive(state, k)
}

func (v *tableView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveTable(v.app.Config.K9s.CurrentCluster, v.GetBaseTitle(), v.GetData()); err != nil {
		v.app.Flash().Err(err)
	} else {
		v.app.Flash().Infof("File %s saved successfully!", path)
	}

	return nil
}

func (v *tableView) setFilterFn(fn func(string)) {
	v.filterFn = fn

	cmd := v.SearchBuff().String()
	if isLabelSelector(cmd) && v.filterFn != nil {
		v.filterFn(trimLabelSelector(cmd))
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
	cmd := v.SearchBuff().String()
	if isLabelSelector(cmd) && v.filterFn != nil {
		v.filterFn(trimLabelSelector(cmd))
		return nil
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
	if isLabelSelector(v.SearchBuff().String()) {
		v.filterFn("")
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
