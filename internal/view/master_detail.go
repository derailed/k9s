package view

import (
	"context"

	"github.com/derailed/k9s/internal/ui"
)

// MasterDetail presents a master-detail viewer.
type MasterDetail struct {
	*PageStack

	enterFn        enterFn
	extraActionsFn func(ui.KeyActions)
	details        *Details
}

// NewMasterDetail returns a new master-detail viewer.
func NewMasterDetail() *MasterDetail {
	return &MasterDetail{
		PageStack: NewPageStack(),
	}
}

// Init initializes the viewer.
func (m *MasterDetail) Init(ctx context.Context) {
	m.PageStack.Init(ctx)

	t := NewTable("master")
	t.Init(ctx)
	m.Push(t)

	m.details = NewDetails(m.app, nil)
	m.details.Init(ctx)
}

func (m *MasterDetail) setExtraActionsFn(f ActionsFunc) {
	m.extraActionsFn = f
}

// Protocol...

func (m *MasterDetail) setEnterFn(f enterFn) {
	m.enterFn = f
}

func (m *MasterDetail) showMaster() {
	m.Show("table")
}

func (m *MasterDetail) masterPage() *Table {
	return m.GetPrimitive("table").(*Table)
}

func (m *MasterDetail) showDetails() {
	m.Push(m.details)
}

func (m *MasterDetail) detailsPage() *Details {
	return m.details
}

// ----------------------------------------------------------------------------
// Actions...

func (m *MasterDetail) defaultActions(aa ui.KeyActions) {
	aa[ui.KeyHelp] = ui.NewKeyAction("Help", noopCmd, false)
	aa[ui.KeyP] = ui.NewKeyAction("Previous", m.app.PrevCmd, false)

	if m.extraActionsFn != nil {
		m.extraActionsFn(aa)
	}
}
