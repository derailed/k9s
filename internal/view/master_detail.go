package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// MasterDetail presents a master-detail viewer.
type MasterDetail struct {
	*PageStack

	enterFn        enterFn
	extraActionsFn func(ui.KeyActions)
	master         *Table
	details        *Details
	currentNS      string
	title          string
}

// NewMasterDetail returns a new master-detail viewer.
func NewMasterDetail(title, ns string) *MasterDetail {
	return &MasterDetail{
		PageStack: NewPageStack(),
		title:     title,
		currentNS: ns,
	}
}

// Init initializes the viewer.
func (m *MasterDetail) Init(ctx context.Context) {
	app := ctx.Value(ui.KeyApp).(*App)
	if m.currentNS != resource.NotNamespaced {
		m.currentNS = app.Config.ActiveNamespace()
	}
	m.PageStack.Init(ctx)
	m.AddListener(app.Menu())

	m.master = NewTable(m.title)
	m.Push(m.master)

	m.details = NewDetails(m.app, func(evt *tcell.EventKey) *tcell.EventKey {
		m.Pop()
		return nil
	})
	m.details.Init(ctx)
}

// Hints returns the current viewer hints
func (m *MasterDetail) Hints() model.MenuHints {
	if c, ok := m.Top().(model.Hinter); ok {
		return c.Hints()
	}

	return nil
}

func (m *MasterDetail) setExtraActionsFn(f ActionsFunc) {
	m.extraActionsFn = f
}

// Protocol...

func (m *MasterDetail) setEnterFn(f enterFn) {
	m.enterFn = f
}

func (m *MasterDetail) showMaster() {
	m.Show(m.master)
}

func (m *MasterDetail) masterPage() *Table {
	return m.master
}

func (m *MasterDetail) showDetails() {
	m.Push(m.details)
}

func (m *MasterDetail) detailsPage() *Details {
	return m.details
}

func (m *MasterDetail) isMaster() bool {
	return m.Current() == m.master
}

// ----------------------------------------------------------------------------
// Actions...

func (m *MasterDetail) defaultActions(aa ui.KeyActions) {
	aa[ui.KeyHelp] = ui.NewKeyAction("Help", noopCmd, false)
	aa[tcell.KeyEsc] = ui.NewKeyAction("Back", m.backCmd, false)

	if m.extraActionsFn != nil {
		m.extraActionsFn(aa)
	}
}

func (m *MasterDetail) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	m.DumpPages()
	m.DumpStack()

	if !m.isMaster() {
		return m.app.PrevCmd(evt)
	}

	if m.masterPage().resetCmd(evt) != nil {
		return m.app.PrevCmd(evt)
	}

	return nil
}

func (m *MasterDetail) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := m.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}
