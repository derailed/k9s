package ui

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

type (
	// ColorerFunc represents a row colorer.
	ColorerFunc func(ns string, evt render.RowEvent) tcell.Color

	// DecorateFunc represents a row decorator.
	DecorateFunc func(render.TableData) render.TableData

	// SelectedRowFunc a table selection callback.
	SelectedRowFunc func(r int)
)

// Table represents tabular data.
type Table struct {
	*SelectTable

	actions    KeyActions
	BaseTitle  string
	Path       string
	cmdBuff    *CmdBuff
	styles     *config.Styles
	sortCol    SortColumn
	colorerFn  render.ColorerFunc
	decorateFn DecorateFunc
}

// NewTable returns a new table view.
func NewTable(gvr string) *Table {
	return &Table{
		SelectTable: &SelectTable{
			Table:       tview.NewTable(),
			model:       model.NewTable(gvr),
			selectedRow: 1,
			marks:       make(map[string]struct{}),
		},
		actions:   make(KeyActions),
		cmdBuff:   NewCmdBuff('/', FilterBuff),
		BaseTitle: gvr,
		sortCol:   SortColumn{index: -1, colCount: 0, asc: true},
	}
}

// Init initializes the component.
func (t *Table) Init(ctx context.Context) {
	t.SetFixed(1, 0)
	t.SetBorder(true)
	t.SetBorderAttributes(tcell.AttrBold)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false)
	t.SetSelectionChangedFunc(t.selectionChanged)
	t.SetInputCapture(t.keyboard)

	t.styles = mustExtractSyles(ctx)
	t.StylesChanged(t.styles)
}

// StylesChanged notifies the skin changed.
func (t *Table) StylesChanged(s *config.Styles) {
	t.SetBackgroundColor(config.AsColor(s.Table().BgColor))
	t.SetBorderColor(config.AsColor(s.Table().FgColor))
	t.SetBorderFocusColor(config.AsColor(s.Frame().Border.FocusColor))
	t.SetSelectedStyle(
		tcell.ColorBlack,
		config.AsColor(t.styles.Table().CursorColor),
		tcell.AttrBold,
	)
	t.Refresh()
}

// Actions returns active menu bindings.
func (t *Table) Actions() KeyActions {
	return t.actions
}

// Styles returns styling configurator.
func (t *Table) Styles() *config.Styles {
	return t.styles
}

// SendKey sends an keyboard event (testing only!).
func (t *Table) SendKey(evt *tcell.EventKey) {
	t.keyboard(evt)
}

func (t *Table) filterInput(r rune) bool {
	if !t.cmdBuff.IsActive() {
		return false
	}
	t.cmdBuff.Add(r)
	t.ClearSelection()
	t.doUpdate(t.filtered(t.GetModel().Peek()))
	t.UpdateTitle()
	t.SelectFirstRow()

	return true
}

func (t *Table) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyUp || key == tcell.KeyDown {
		return evt
	}

	if key == tcell.KeyRune {
		if t.filterInput(evt.Rune()) {
			return nil
		}
		key = AsKey(evt)
	}

	if a, ok := t.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

// Hints returns the view hints.
func (t *Table) Hints() model.MenuHints {
	return t.actions.Hints()
}

// ExtraHints returns additional hints.
func (t *Table) ExtraHints() map[string]string {
	return nil
}

// GetFilteredData fetch filtered tabular data.
func (t *Table) GetFilteredData() render.TableData {
	return t.filtered(t.GetModel().Peek())
}

// SetDecorateFn specifies the default row decorator.
func (t *Table) SetDecorateFn(f DecorateFunc) {
	t.decorateFn = f
}

// SetColorerFn specifies the default colorer.
func (t *Table) SetColorerFn(f render.ColorerFunc) {
	t.colorerFn = f
}

// SetSortCol sets in sort column index and order.
func (t *Table) SetSortCol(index, count int, asc bool) {
	t.sortCol.index, t.sortCol.colCount, t.sortCol.asc = index, count, asc
}

// Update table content.
func (t *Table) Update(data render.TableData) {
	if t.decorateFn != nil {
		data = t.decorateFn(data)
	}
	if !t.cmdBuff.Empty() {
		data = t.filtered(data)
	}
	t.doUpdate(data)
	t.UpdateTitle()
}

func (t *Table) doUpdate(data render.TableData) {
	if client.IsAllNamespaces(data.Namespace) {
		t.actions[KeyShiftP] = NewKeyAction("Sort Namespace", t.SortColCmd(-2, true), false)
	} else {
		t.actions.Delete(KeyShiftP)
	}

	t.Clear()
	t.adjustSorter(data)
	fg := config.AsColor(t.styles.Table().Header.FgColor)
	bg := config.AsColor(t.styles.Table().Header.BgColor)
	for col, h := range data.Header {
		t.AddHeaderCell(col, h)
		c := t.GetCell(0, col)
		c.SetBackgroundColor(bg)
		c.SetTextColor(fg)
	}
	data.RowEvents.Sort(data.Namespace, t.sortCol.index, t.sortCol.asc)

	pads := make(MaxyPad, len(data.Header))
	ComputeMaxColumns(pads, t.sortCol.index, data.Header, data.RowEvents)
	for i, r := range data.RowEvents {
		t.buildRow(data.Namespace, i+1, r, data.Header, pads)
	}
	t.updateSelection(true)
}

// SortColCmd designates a sorted column.
func (t *Table) SortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		var index int
		switch col {
		case -2:
			index = 0
		case -1:
			index = t.GetColumnCount() - 1
		default:
			index = t.NameColIndex() + col
		}
		t.sortCol.asc = !t.sortCol.asc
		if t.sortCol.index != index {
			t.sortCol.asc = asc
		}
		t.sortCol.index = index
		t.Refresh()
		return nil
	}
}

// SortInvertCmd reverses sorting order.
func (t *Table) SortInvertCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.sortCol.asc = !t.sortCol.asc
	t.Refresh()

	return nil
}

func (t *Table) adjustSorter(data render.TableData) {
	// Going from namespace to non namespace or vice-versa?
	switch {
	case t.sortCol.colCount == 0:
	case len(data.Header) > t.sortCol.colCount:
		t.sortCol.index++
	case len(data.Header) < t.sortCol.colCount:
		t.sortCol.index--
	}
	t.sortCol.colCount = len(data.Header)
	if t.sortCol.index < 0 {
		t.sortCol.index = 0
	}
}

func (t *Table) buildRow(ns string, r int, re render.RowEvent, header render.HeaderRow, pads MaxyPad) {
	color := render.DefaultColorer
	if t.colorerFn != nil {
		color = t.colorerFn
	}
	marked := t.IsMarked(re.Row.ID)
	for col, field := range re.Row.Fields {
		if !re.Deltas.IsBlank() && !header.AgeCol(col) {
			field += Deltas(re.Deltas[col], field)
		}

		if header[col].Decorator != nil {
			field = header[col].Decorator(field)
		}

		if header[col].Align == tview.AlignLeft {
			field = formatCell(field, pads[col])
		}
		c := tview.NewTableCell(field)
		c.SetExpansion(1)
		c.SetAlign(header[col].Align)
		c.SetTextColor(color(ns, re))
		if marked {
			c.SetTextColor(config.AsColor(t.styles.Table().MarkColor))
		}
		if col == 0 {
			c.SetReference(re.Row.ID)
		}
		t.SetCell(r, col, c)
	}
}

// ClearMarks clear out marked items.
func (t *Table) ClearMarks() {
	t.SelectTable.ClearMarks()
	t.Refresh()
}

// Refresh update the table data.
func (t *Table) Refresh() {
	// BOZO!! Really want to tell model reload now. Refactor!
	t.Update(t.model.Peek())
}

// GetSelectedRow returns the entire selected row.
func (t *Table) GetSelectedRow() render.Row {
	log.Debug().Msgf("INDEX %d", t.GetSelectedRowIndex())
	return t.model.Peek().RowEvents[t.GetSelectedRowIndex()-1].Row
}

// NameColIndex returns the index of the resource name column.
func (t *Table) NameColIndex() int {
	col := 0
	if t.GetModel().ClusterWide() {
		col++
	}
	return col
}

// AddHeaderCell configures a table cell header.
func (t *Table) AddHeaderCell(col int, h render.Header) {
	c := tview.NewTableCell(sortIndicator(t.sortCol, t.styles.Table(), col, h.Name))
	c.SetExpansion(1)
	c.SetAlign(h.Align)
	t.SetCell(0, col, c)
}

func (t *Table) filtered(data render.TableData) render.TableData {
	if t.cmdBuff.Empty() || IsLabelSelector(t.cmdBuff.String()) {
		return data
	}
	q := t.cmdBuff.String()
	if IsFuzzySelector(q) {
		return fuzzyFilter(q[2:], t.NameColIndex(), data)
	}

	filtered, err := rxFilter(t.cmdBuff.String(), data)
	if err != nil {
		log.Error().Err(errors.New("Invalid filter expression")).Msg("Regexp")
		t.cmdBuff.Clear()
		return data
	}
	return filtered
}

// SearchBuff returns the associated command buffer.
func (t *Table) SearchBuff() *CmdBuff {
	return t.cmdBuff
}

// ShowDeleted marks row as deleted.
func (t *Table) ShowDeleted() {
	r, _ := t.GetSelection()
	cols := t.GetColumnCount()
	for x := 0; x < cols; x++ {
		t.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// UpdateTitle refreshes the table title.
func (t *Table) UpdateTitle() {
	t.SetTitle(t.styleTitle())
}

func (t *Table) styleTitle() string {
	rc := t.GetRowCount()
	if rc > 0 {
		rc--
	}

	base := strings.Title(t.BaseTitle)
	ns := t.GetModel().GetNamespace()
	if client.IsAllNamespaces(ns) {
		ns = client.NamespaceAll
	}
	path := t.Path
	if path != "" {
		cns, n := client.Namespaced(path)
		if cns == client.ClusterScope {
			ns = n
		} else {
			ns = path
		}
	}

	var title string
	if ns == client.ClusterScope {
		title = SkinTitle(fmt.Sprintf(TitleFmt, base, rc), t.styles.Frame())
	} else {
		title = SkinTitle(fmt.Sprintf(NSTitleFmt, base, ns, rc), t.styles.Frame())
	}

	buff := t.cmdBuff.String()
	if buff == "" {
		return title
	}
	if IsLabelSelector(buff) {
		buff = TrimLabelSelector(buff)
	}

	return title + SkinTitle(fmt.Sprintf(SearchFmt, buff), t.styles.Frame())
}
