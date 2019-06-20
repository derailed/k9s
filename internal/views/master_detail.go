package views

import (
	"path"
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
)

type (
	pageView struct {
		*tview.Pages

		app          *appView
		actions      keyActions
		selectedItem string
		selectedRow  int
		selectedFn   func() string
	}

	masterDetail struct {
		*pageView

		currentNS      string
		enterFn        enterFn
		extraActionsFn func(keyActions)
	}
)

func newPageView(app *appView) *pageView {
	return &pageView{
		Pages:   tview.NewPages(),
		app:     app,
		actions: make(keyActions),
	}
}

func newMasterDetail(title string, app *appView, ns string) *masterDetail {
	v := masterDetail{
		pageView:  newPageView(app),
		currentNS: ns,
	}

	tv := newTableView(app, title)
	tv.SetSelectionChangedFunc(v.selChanged)
	v.AddPage("master", tv, true, true)

	return &v
}

func (v *masterDetail) init(ns string, backCmd actionHandler) {
	details := newDetailsView(v.app, backCmd)
	v.AddPage("details", details, true, false)

	if v.currentNS != resource.NotNamespaced {
		v.currentNS = ns
	}
}

func (v *masterDetail) setExtraActionsFn(f actionsFn) {
	f(v.actions)
}

func (v *masterDetail) rowSelected() bool {
	return v.selectedItem != ""
}

func (v *masterDetail) selChanged(r, c int) {
	v.selectedRow = r
	v.selectItem(r, c)
}

func (v *masterDetail) getSelectedItem() string {
	if v.selectedFn != nil {
		return v.selectedFn()
	}
	return v.selectedItem
}

// Protocol...

// Hints fetch menu hints
func (v *masterDetail) hints() hints {
	return v.CurrentPage().Item.(hinter).hints()
}

func (v *masterDetail) setEnterFn(f enterFn) {
	v.enterFn = f
}

func (v *masterDetail) masterPage() *tableView {
	return v.GetPrimitive("master").(*tableView)
}

func (v *masterDetail) detailsPage() *detailsView {
	return v.GetPrimitive("details").(*detailsView)
}

// ----------------------------------------------------------------------------
// Actions...

func (v *masterDetail) selectItem(r, c int) {
	t := v.masterPage()
	if r == 0 || t.GetCell(r, 0) == nil {
		v.selectedItem = ""
		return
	}

	col0 := cleanCell(t, r, 0)
	switch v.currentNS {
	case resource.NotNamespaced:
		v.selectedItem = col0
	case resource.AllNamespace, resource.AllNamespaces:
		v.selectedItem = path.Join(col0, cleanCell(t, r, 1))
	default:
		v.selectedItem = path.Join(v.currentNS, col0)
	}
}

func (v *masterDetail) defaultActions() {
	v.actions[KeyHelp] = newKeyAction("Help", noopCmd, false)
	v.actions[KeyP] = newKeyAction("Previous", v.app.prevCmd, false)

	if v.extraActionsFn != nil {
		v.extraActionsFn(v.actions)
	}
}

func cleanCell(v *tableView, r, c int) string {
	return strings.TrimSpace(v.GetCell(r, c).Text)
}
