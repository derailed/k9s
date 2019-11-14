package ui

import (
	"context"
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"github.com/sahilm/fuzzy"
	"k8s.io/apimachinery/pkg/util/duration"
)

type (
	// ColorerFunc represents a row colorer.
	ColorerFunc func(ns string, evt *resource.RowEvent) tcell.Color

	// SelectedRowFunc a table selection callback.
	SelectedRowFunc func(r, c int)
)

// Table represents tabular data.
type Table struct {
	*tview.Table

	baseTitle    string
	data         resource.TableData
	actions      KeyActions
	cmdBuff      *CmdBuff
	styles       *config.Styles
	activeNS     string
	sortCol      SortColumn
	sortFn       SortFn
	colorerFn    ColorerFunc
	selectedItem string
	selectedRow  int
	selectedFn   func(string) string
	selListeners []SelectedRowFunc
	marks        map[string]bool
}

// NewTable returns a new table view.
func NewTable(title string) *Table {
	return &Table{
		Table:     tview.NewTable(),
		actions:   make(KeyActions),
		cmdBuff:   NewCmdBuff('/', FilterBuff),
		baseTitle: title,
		sortCol:   SortColumn{0, 0, true},
		marks:     make(map[string]bool),
	}
}

func (t *Table) Init(ctx context.Context) {
	log.Debug().Msgf("UI Table INIT %q", t.baseTitle)
	t.styles = ctx.Value(KeyStyles).(*config.Styles)

	t.SetFixed(1, 0)
	t.SetBorder(true)
	t.SetBackgroundColor(config.AsColor(t.styles.Table().BgColor))
	t.SetBorderColor(config.AsColor(t.styles.Table().FgColor))
	t.SetBorderFocusColor(config.AsColor(t.styles.Frame().Border.FocusColor))
	t.SetBorderAttributes(tcell.AttrBold)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false)
	t.SetSelectedStyle(
		tcell.ColorBlack,
		config.AsColor(t.styles.Table().CursorColor),
		tcell.AttrBold,
	)

	t.SetSelectionChangedFunc(t.selChanged)
	t.SetInputCapture(t.keyboard)
}

// SendKey sends an keyboard event (testing only!).
func (t *Table) SendKey(evt *tcell.EventKey) {
	t.keyboard(evt)
}

// GetRow retrieves the entire selected row.
func (t *Table) GetRow() resource.Row {
	r := make(resource.Row, t.GetColumnCount())
	for i := 0; i < t.GetColumnCount(); i++ {
		c := t.GetCell(t.selectedRow, i)
		r[i] = strings.TrimSpace(c.Text)
	}
	return r
}

// AddSelectedRowListener add a new selected row listener.
func (t *Table) AddSelectedRowListener(f SelectedRowFunc) {
	t.selListeners = append(t.selListeners, f)
}

func (t *Table) selChanged(r, c int) {
	t.selectedRow = r
	t.updateSelectedItem(r)
	if r == 0 {
		return
	}

	cell := t.GetCell(r, c)
	t.SetSelectedStyle(
		tcell.ColorBlack,
		cell.Color,
		tcell.AttrBold,
	)

	for _, f := range t.selListeners {
		f(r, c)
	}
}

// UpdateSelection refresh selected row.
func (t *Table) updateSelection(broadcast bool) {
	t.SelectRow(t.selectedRow, broadcast)
}

// SelectRow select a given row by index.
func (t *Table) SelectRow(r int, broadcast bool) {
	if !broadcast {
		t.SetSelectionChangedFunc(nil)
	}
	defer t.SetSelectionChangedFunc(t.selChanged)
	t.Select(r, 0)
	t.updateSelectedItem(r)
}

func (t *Table) updateSelectedItem(r int) {
	if r == 0 || t.GetCell(r, 0) == nil {
		t.selectedItem = ""
		return
	}

	col0 := TrimCell(t, r, 0)
	switch t.activeNS {
	case resource.NotNamespaced:
		t.selectedItem = col0
	case resource.AllNamespace, resource.AllNamespaces:
		t.selectedItem = path.Join(col0, TrimCell(t, r, 1))
	default:
		t.selectedItem = path.Join(t.activeNS, col0)
	}
}

// SetSelectedFn defines a function that cleanse the current selection.
func (t *Table) SetSelectedFn(f func(string) string) {
	t.selectedFn = f
}

// RowSelected checks if there is an active row selection.
func (t *Table) RowSelected() bool {
	return t.selectedItem != ""
}

// GetSelectedCell returns the content of a cell for the currently selected row.
func (t *Table) GetSelectedCell(col int) string {
	return TrimCell(t, t.selectedRow, col)
}

// GetSelectedRow fetch the currently selected row index.
func (t *Table) GetSelectedRowIndex() int {
	return t.selectedRow
}

// GetSelectedItem returns the currently selected item name.
func (t *Table) GetSelectedItem() string {
	if t.selectedFn != nil {
		return t.selectedFn(t.selectedItem)
	}
	return t.selectedItem
}

// GetSelectedItems return currently marked or selected items names.
func (t *Table) GetSelectedItems() []string {
	if len(t.marks) > 0 {
		var items []string
		for item, marked := range t.marks {
			if marked {
				items = append(items, item)
			}
		}
		return items
	}
	return []string{t.GetSelectedItem()}
}

func (t *Table) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if t.SearchBuff().IsActive() {
			t.SearchBuff().Add(evt.Rune())
			t.ClearSelection()
			t.doUpdate(t.filtered())
			t.UpdateTitle()
			t.SelectFirstRow()
			return nil
		}
		key = asKey(evt)
	}

	if a, ok := t.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

// GetData fetch tabular data.
func (t *Table) GetData() resource.TableData {
	return t.data
}

// GetFilteredData fetch filtered tabular data.
func (t *Table) GetFilteredData() resource.TableData {
	return t.filtered()
}

// SetBaseTitle set the table title.
func (t *Table) SetBaseTitle(s string) {
	t.baseTitle = s
}

// GetBaseTitle fetch the current title.
func (t *Table) GetBaseTitle() string {
	return t.baseTitle
}

// SetColorerFn set the row colorer.
func (t *Table) SetColorerFn(f ColorerFunc) {
	t.colorerFn = f
}

// ActiveNS get the resource namespace.
func (t *Table) ActiveNS() string {
	return t.activeNS
}

// SetActiveNS set the resource namespace.
func (t *Table) SetActiveNS(ns string) {
	t.activeNS = ns
}

// SetSortCol sets in sort column index and order.
func (t *Table) SetSortCol(index, count int, asc bool) {
	t.sortCol.index, t.sortCol.colCount, t.sortCol.asc = index, count, asc
}

// Update table content.
func (t *Table) Update(data resource.TableData) {
	t.data = data
	if t.cmdBuff.Empty() {
		t.doUpdate(t.data)
	} else {
		t.doUpdate(t.filtered())
	}
	t.UpdateTitle()
	t.updateSelection(true)
}

func (t *Table) doUpdate(data resource.TableData) {
	t.activeNS = data.Namespace
	if t.activeNS == resource.AllNamespaces && t.activeNS != "*" {
		t.actions[KeyShiftP] = NewKeyAction("Sort Namespace", t.SortColCmd(-2), false)
	} else {
		delete(t.actions, KeyShiftP)
	}
	t.Clear()

	t.adjustSorter(data)

	var row int
	fg := config.AsColor(t.styles.Table().Header.FgColor)
	bg := config.AsColor(t.styles.Table().Header.BgColor)
	for col, h := range data.Header {
		t.AddHeaderCell(data.NumCols[h], col, h)
		c := t.GetCell(0, col)
		c.SetBackgroundColor(bg)
		c.SetTextColor(fg)
	}
	row++

	t.sort(data, row)
}

// SortColCmd designates a sorted column.
func (t *Table) SortColCmd(col int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t.sortCol.asc = true
		switch col {
		case -2:
			t.sortCol.index = 0
		case -1:
			t.sortCol.index = t.GetColumnCount() - 1
		default:
			t.sortCol.index = t.NameColIndex() + col

		}
		t.Refresh()

		return nil
	}
}

func (t *Table) adjustSorter(data resource.TableData) {
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

func (t *Table) sort(data resource.TableData, row int) {
	pads := make(MaxyPad, len(data.Header))
	ComputeMaxColumns(pads, t.sortCol.index, data)

	sortFn := defaultSort
	if t.sortFn != nil {
		sortFn = t.sortFn
	}
	prim, sec := sortAllRows(t.sortCol, data.Rows, sortFn)
	for _, pk := range prim {
		for _, sk := range sec[pk] {
			t.buildRow(row, data, sk, pads)
			row++
		}
	}
}

func (t *Table) buildRow(row int, data resource.TableData, sk string, pads MaxyPad) {
	f := DefaultColorer
	if t.colorerFn != nil {
		f = t.colorerFn
	}
	m := t.isMarked(sk)
	for col, field := range data.Rows[sk].Fields {
		header := data.Header[col]
		field, align := t.formatCell(data.NumCols[header], header, field+Deltas(data.Rows[sk].Deltas[col], field), pads[col])
		c := tview.NewTableCell(field)
		{
			c.SetExpansion(1)
			c.SetAlign(align)
			c.SetTextColor(f(data.Namespace, data.Rows[sk]))
			if m {
				c.SetBackgroundColor(config.AsColor(t.styles.Table().MarkColor))
			}
		}
		t.SetCell(row, col, c)
	}
}

func (t *Table) formatCell(numerical bool, header, field string, padding int) (string, int) {
	if header == "AGE" {
		dur, err := time.ParseDuration(field)
		if err == nil {
			field = duration.HumanDuration(dur)
		}
	}

	if numerical || cpuRX.MatchString(header) || memRX.MatchString(header) {
		return field, tview.AlignRight
	}

	align := tview.AlignLeft
	if IsASCII(field) {
		return Pad(field, padding), align
	}

	return field, align
}

// Refresh update the table data.
func (t *Table) Refresh() {
	t.Update(t.data)
}

// NameColIndex returns the index of the resource name column.
func (t *Table) NameColIndex() int {
	col := 0
	if t.activeNS == resource.AllNamespaces {
		col++
	}
	return col
}

// AddHeaderCell configures a table cell header.
func (t *Table) AddHeaderCell(numerical bool, col int, name string) {
	c := tview.NewTableCell(sortIndicator(t.sortCol, t.styles.Table(), col, name))
	c.SetExpansion(1)
	if numerical || cpuRX.MatchString(name) || memRX.MatchString(name) {
		c.SetAlign(tview.AlignRight)
	}
	t.SetCell(0, col, c)
}

func (t *Table) filtered() resource.TableData {
	if t.cmdBuff.Empty() || IsLabelSelector(t.cmdBuff.String()) {
		return t.data
	}

	q := t.cmdBuff.String()
	if isFuzzySelector(q) {
		return t.fuzzyFilter(q[2:])
	}

	return t.rxFilter(q)
}

func (t *Table) rxFilter(q string) resource.TableData {
	rx, err := regexp.Compile(`(?i)` + t.cmdBuff.String())
	if err != nil {
		log.Error().Err(errors.New("Invalid filter expression")).Msg("Regexp")
		t.cmdBuff.Clear()
		return t.data
	}

	filtered := resource.TableData{
		Header:    t.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: t.data.Namespace,
	}
	for k, row := range t.data.Rows {
		f := strings.Join(row.Fields, " ")
		if rx.MatchString(f) {
			filtered.Rows[k] = row
		}
	}

	return filtered
}

func (t *Table) fuzzyFilter(q string) resource.TableData {
	var ss, kk []string
	for k, row := range t.data.Rows {
		ss = append(ss, row.Fields[t.NameColIndex()])
		kk = append(kk, k)
	}

	filtered := resource.TableData{
		Header:    t.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: t.data.Namespace,
	}
	mm := fuzzy.Find(q, ss)
	for _, m := range mm {
		filtered.Rows[kk[m.Index]] = t.data.Rows[kk[m.Index]]
	}

	return filtered
}

// KeyBindings returns the bounded keys.
func (t *Table) KeyBindings() KeyActions {
	return t.actions
}

// SearchBuff returns the associated command buffer.
func (t *Table) SearchBuff() *CmdBuff {
	return t.cmdBuff
}

// ClearSelection reset selected row.
func (t *Table) ClearSelection() {
	t.Select(0, 0)
	t.ScrollToBeginning()
}

// SelectFirstRow select first data row if any.
func (t *Table) SelectFirstRow() {
	if t.GetRowCount() > 0 {
		t.Select(1, 0)
	}
}

// ShowDeleted marks row as deleted.
func (t *Table) ShowDeleted() {
	r, _ := t.GetSelection()
	cols := t.GetColumnCount()
	for x := 0; x < cols; x++ {
		t.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// SetActions sets up keyboard action listener.
func (t *Table) AddActions(aa KeyActions) {
	for k, a := range aa {
		t.actions[k] = a
	}
}

// RmAction delete a keyed action.
func (t *Table) RmAction(kk ...tcell.Key) {
	for _, k := range kk {
		delete(t.actions, k)
	}
}

// Hints options
func (t *Table) Hints() model.MenuHints {
	if t.actions != nil {
		return t.actions.Hints()
	}

	return nil
}

// UpdateTitle refreshes the table title.
func (t *Table) UpdateTitle() {
	var title string

	rc := t.GetRowCount()
	if rc > 0 {
		rc--
	}
	switch t.activeNS {
	case resource.NotNamespaced, "*":
		title = skinTitle(fmt.Sprintf(titleFmt, t.baseTitle, rc), t.styles.Frame())
	default:
		ns := t.activeNS
		if ns == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = skinTitle(fmt.Sprintf(nsTitleFmt, t.baseTitle, ns, rc), t.styles.Frame())
	}

	if !t.cmdBuff.Empty() {
		cmd := t.cmdBuff.String()
		if IsLabelSelector(cmd) {
			cmd = TrimLabelSelector(cmd)
		}
		title += skinTitle(fmt.Sprintf(SearchFmt, cmd), t.styles.Frame())
	}
	t.SetTitle(title)
}

// SortInvertCmd reverses sorting order.
func (t *Table) SortInvertCmd(evt *tcell.EventKey) *tcell.EventKey {
	t.sortCol.asc = !t.sortCol.asc
	t.Refresh()

	return nil
}

// ToggleMark toggles marked row
func (t *Table) ToggleMark() {
	t.marks[t.GetSelectedItem()] = !t.marks[t.GetSelectedItem()]
}

func (t *Table) isMarked(item string) bool {
	return t.marks[item]
}
