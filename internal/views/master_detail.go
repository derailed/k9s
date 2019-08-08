package views

import (
	"context"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
)

type (
	pageView struct {
		*tview.Pages

		app *appView
	}

	masterDetail struct {
		*pageView

		currentNS      string
		title          string
		enterFn        enterFn
		extraActionsFn func(ui.KeyActions)
	}
)

func newPageView(app *appView) *pageView {
	return &pageView{
		Pages: tview.NewPages(),
		app:   app,
	}
}

func newMasterDetail(title, ns string, app *appView, backCmd ui.ActionHandler) *masterDetail {
	v := masterDetail{
		pageView:  newPageView(app),
		currentNS: ns,
		title:     title,
	}
	v.AddPage("master", newTableView(v.app, v.title), true, true)
	v.AddPage("details", newDetailsView(v.app, backCmd), true, false)

	return &v
}

func (v *masterDetail) init(ctx context.Context, ns string) {
	if v.currentNS != resource.NotNamespaced {
		v.currentNS = ns
	}
}

func (v *masterDetail) setExtraActionsFn(f ui.ActionsFunc) {
	v.extraActionsFn = f
	// f(v.actions)
}

// Protocol...

// Hints fetch menu hints
func (v *masterDetail) hints() ui.Hints {
	return v.CurrentPage().Item.(ui.Hinter).Hints()
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

func (v *masterDetail) defaultActions(aa ui.KeyActions) {
	aa[ui.KeyHelp] = ui.NewKeyAction("Help", noopCmd, false)
	aa[ui.KeyP] = ui.NewKeyAction("Previous", v.app.prevCmd, false)

	if v.extraActionsFn != nil {
		v.extraActionsFn(aa)
	}
}
