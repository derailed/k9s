package view

import (
	"context"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// MasterDetail presents a master-detail viewer.
type MasterDetail struct {
	*PageStack

	enterFn        enterFn
	extraActionsFn func(ui.KeyActions)
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
	log.Debug().Msgf("\t>>>MasterDetail init %q", m.title)
	app := ctx.Value(ui.KeyApp).(*App)
	if m.currentNS != resource.NotNamespaced {
		m.currentNS = app.Config.ActiveNamespace()
	}
	m.PageStack.Init(ctx)
	m.AddListener(app.Menu())

	t := NewTable(m.title)
	m.Push(t)

	m.details = NewDetails(m.app, func(evt *tcell.EventKey) *tcell.EventKey {
		m.Pop()
		return nil
	})
	m.details.Init(ctx)
	log.Debug().Msgf("\t<<<<MasterDetail INIT DONE!!")
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
	m.Show("table")
}

func (m *MasterDetail) masterPage() *Table {
	return m.GetPrimitive(m.title).(*Table)
}

func (m *MasterDetail) showDetails() {
	m.Push(m.details)
}

func (m *MasterDetail) detailsPage() *Details {
	return m.details
}

func (m *MasterDetail) isMaster() bool {
	p := m.CurrentPage()
	if p == nil {
		return false
	}

	log.Debug().Msgf("!!!!!Checking MASTER %q vs %q -- %t", p.Name, m.title, p.Name == m.title)
	return p.Name == m.title
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
