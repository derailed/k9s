package ui

import (
	"errors"
	"fmt"
	"path"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
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
}

// NewTable returns a new table view.
func NewTable(title string, styles *config.Styles) *Table {
	v := Table{
		Table:     tview.NewTable(),
		styles:    styles,
		actions:   make(KeyActions),
		cmdBuff:   NewCmdBuff('/', FilterBuff),
		baseTitle: title,
		sortCol:   SortColumn{0, 0, true},
	}

	v.SetFixed(1, 0)
	v.SetBorder(true)
	v.SetBackgroundColor(config.AsColor(styles.Table().BgColor))
	v.SetBorderColor(config.AsColor(styles.Table().FgColor))
	v.SetBorderFocusColor(config.AsColor(styles.Frame().Border.FocusColor))
	v.SetBorderAttributes(tcell.AttrBold)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetSelectable(true, false)
	v.SetSelectedStyle(
		tcell.ColorBlack,
		config.AsColor(styles.Table().CursorColor),
		tcell.AttrBold,
	)

	v.SetSelectionChangedFunc(v.selChanged)
	v.SetInputCapture(v.keyboard)

	return &v
}

// GetRow retrieves the entire selected row.
func (v *Table) GetRow() resource.Row {
	r := make(resource.Row, v.GetColumnCount())
	for i := 0; i < v.GetColumnCount(); i++ {
		c := v.GetCell(v.selectedRow, i)
		r[i] = strings.TrimSpace(c.Text)
	}
	return r
}

// AddSelectedRowListener add a new selected row listener.
func (v *Table) AddSelectedRowListener(f SelectedRowFunc) {
	v.selListeners = append(v.selListeners, f)
}

func (v *Table) selChanged(r, c int) {
	v.selectedRow = r
	v.updateSelectedItem(r)
	if r == 0 {
		return
	}

	cell := v.GetCell(r, c)
	v.SetSelectedStyle(
		tcell.ColorBlack,
		cell.Color,
		tcell.AttrBold,
	)

	for _, f := range v.selListeners {
		f(r, c)
	}
}

// UpdateSelection refresh selected row.
func (v *Table) updateSelection(broadcast bool) {
	v.SelectRow(v.selectedRow, broadcast)
}

// SelectRow select a given row by index.
func (v *Table) SelectRow(r int, broadcast bool) {
	if !broadcast {
		v.SetSelectionChangedFunc(nil)
	}
	defer v.SetSelectionChangedFunc(v.selChanged)
	v.Select(r, 0)
	v.updateSelectedItem(r)
}

func (v *Table) updateSelectedItem(r int) {
	if r == 0 || v.GetCell(r, 0) == nil {
		v.selectedItem = ""
		return
	}

	col0 := TrimCell(v, r, 0)
	switch v.activeNS {
	case resource.NotNamespaced:
		v.selectedItem = col0
	case resource.AllNamespace, resource.AllNamespaces:
		v.selectedItem = path.Join(col0, TrimCell(v, r, 1))
	default:
		v.selectedItem = path.Join(v.activeNS, col0)
	}
}

// SetSelectedFn defines a function that cleanse the current selection.
func (v *Table) SetSelectedFn(f func(string) string) {
	v.selectedFn = f
}

// RowSelected checks if there is an active row selection.
func (v *Table) RowSelected() bool {
	return v.selectedItem != ""
}

// GetSelectedCell returns the contant of a cell for the currently selected row.
func (v *Table) GetSelectedCell(col int) string {
	return TrimCell(v, v.selectedRow, col)
}

// GetSelectedRow fetch the currently selected row index.
func (v *Table) GetSelectedRow() int {
	return v.selectedRow
}

// GetSelectedItem returns the currently selected item name.
func (v *Table) GetSelectedItem() string {
	if v.selectedFn != nil {
		return v.selectedFn(v.selectedItem)
	}
	return v.selectedItem
}

func (v *Table) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if v.SearchBuff().IsActive() {
			v.SearchBuff().Add(evt.Rune())
			v.ClearSelection()
			v.doUpdate(v.filtered())
			v.UpdateTitle()
			v.SelectFirstRow()
			return nil
		}
		key = asKey(evt)
	}

	if a, ok := v.actions[key]; ok {
		return a.Action(evt)
	}

	return evt
}

// GetData fetch tabular data.
func (v *Table) GetData() resource.TableData {
	return v.data
}

// GetFilteredData fetch filtered tabular data.
func (v *Table) GetFilteredData() resource.TableData {
	return v.filtered()
}

// SetBaseTitle set the table title.
func (v *Table) SetBaseTitle(s string) {
	v.baseTitle = s
}

// GetBaseTitle fetch the current title.
func (v *Table) GetBaseTitle() string {
	return v.baseTitle
}

// SetColorerFn set the row colorer.
func (v *Table) SetColorerFn(f ColorerFunc) {
	v.colorerFn = f
}

// ActiveNS get the resource namespace.
func (v *Table) ActiveNS() string {
	return v.activeNS
}

// SetActiveNS set the resource namespace.
func (v *Table) SetActiveNS(ns string) {
	v.activeNS = ns
}

// SetSortCol sets in sort column index and order.
func (v *Table) SetSortCol(index, count int, asc bool) {
	v.sortCol.index, v.sortCol.colCount, v.sortCol.asc = index, count, asc
}

// Update table content.
func (v *Table) Update(data resource.TableData) {
	v.data = data
	if v.cmdBuff.Empty() {
		v.doUpdate(v.data)
	} else {
		v.doUpdate(v.filtered())
	}
	v.UpdateTitle()
	v.updateSelection(true)
}

func (v *Table) doUpdate(data resource.TableData) {
	v.activeNS = data.Namespace
	if v.activeNS == resource.AllNamespaces && v.activeNS != "*" {
		v.actions[KeyShiftP] = NewKeyAction("Sort Namespace", v.SortColCmd(-2), false)
	} else {
		delete(v.actions, KeyShiftP)
	}
	v.Clear()

	v.adjustSorter(data)

	var row int
	fg := config.AsColor(v.styles.Table().Header.FgColor)
	bg := config.AsColor(v.styles.Table().Header.BgColor)
	for col, h := range data.Header {
		v.AddHeaderCell(data.NumCols[h], col, h)
		c := v.GetCell(0, col)
		c.SetBackgroundColor(bg)
		c.SetTextColor(fg)
	}
	row++

	v.sort(data, row)
}

// SortColCmd designates a sorted column.
func (v *Table) SortColCmd(col int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		v.sortCol.asc = true
		switch col {
		case -2:
			v.sortCol.index = 0
		case -1:
			v.sortCol.index = v.GetColumnCount() - 1
		default:
			v.sortCol.index = v.NameColIndex() + col

		}
		v.Refresh()

		return nil
	}
}

func (v *Table) adjustSorter(data resource.TableData) {
	// Going from namespace to non namespace or vice-versa?
	switch {
	case v.sortCol.colCount == 0:
	case len(data.Header) > v.sortCol.colCount:
		v.sortCol.index++
	case len(data.Header) < v.sortCol.colCount:
		v.sortCol.index--
	}
	v.sortCol.colCount = len(data.Header)
	if v.sortCol.index < 0 {
		v.sortCol.index = 0
	}
}

func (v *Table) sort(data resource.TableData, row int) {
	pads := make(MaxyPad, len(data.Header))
	ComputeMaxColumns(pads, v.sortCol.index, data)

	sortFn := defaultSort
	if v.sortFn != nil {
		sortFn = v.sortFn
	}
	prim, sec := sortAllRows(v.sortCol, data.Rows, sortFn)
	for _, pk := range prim {
		for _, sk := range sec[pk] {
			v.buildRow(row, data, sk, pads)
			row++
		}
	}
}

func (v *Table) buildRow(row int, data resource.TableData, sk string, pads MaxyPad) {
	f := DefaultColorer
	if v.colorerFn != nil {
		f = v.colorerFn
	}
	for col, field := range data.Rows[sk].Fields {
		header := data.Header[col]
		field, align := v.formatCell(data.NumCols[header], header, field+Deltas(data.Rows[sk].Deltas[col], field), pads[col])
		c := tview.NewTableCell(field)
		{
			c.SetExpansion(1)
			c.SetAlign(align)
			c.SetTextColor(f(data.Namespace, data.Rows[sk]))
		}
		v.SetCell(row, col, c)
	}
}

func (v *Table) formatCell(numerical bool, header, field string, padding int) (string, int) {
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
func (v *Table) Refresh() {
	v.Update(v.data)
}

// NameColIndex returns the index of the resource name column.
func (v *Table) NameColIndex() int {
	col := 0
	if v.activeNS == resource.AllNamespaces {
		col++
	}
	return col
}

// AddHeaderCell configures a table cell header.
func (v *Table) AddHeaderCell(numerical bool, col int, name string) {
	c := tview.NewTableCell(sortIndicator(v.sortCol, v.styles.Table(), col, name))
	c.SetExpansion(1)
	if numerical || cpuRX.MatchString(name) || memRX.MatchString(name) {
		c.SetAlign(tview.AlignRight)
	}
	v.SetCell(0, col, c)
}

func (v *Table) filtered() resource.TableData {
	if v.cmdBuff.Empty() || isLabelSelector(v.cmdBuff.String()) {
		return v.data
	}

	q := v.cmdBuff.String()
	if isFuzzySelector(q) {
		return v.fuzzyFilter(q[2:])
	}

	return v.rxFilter(q)
}

func (v *Table) rxFilter(q string) resource.TableData {
	rx, err := regexp.Compile(`(?i)` + v.cmdBuff.String())
	if err != nil {
		log.Error().Err(errors.New("Invalid filter expression")).Msg("Regexp")
		v.cmdBuff.Clear()
		return v.data
	}

	filtered := resource.TableData{
		Header:    v.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: v.data.Namespace,
	}
	for k, row := range v.data.Rows {
		f := strings.Join(row.Fields, " ")
		if rx.MatchString(f) {
			filtered.Rows[k] = row
		}
	}

	return filtered
}

func (v *Table) fuzzyFilter(q string) resource.TableData {
	var ss, kk []string
	for k, row := range v.data.Rows {
		ss = append(ss, row.Fields[v.NameColIndex()])
		kk = append(kk, k)
	}

	filtered := resource.TableData{
		Header:    v.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: v.data.Namespace,
	}
	mm := fuzzy.Find(q, ss)
	for _, m := range mm {
		filtered.Rows[kk[m.Index]] = v.data.Rows[kk[m.Index]]
	}

	return filtered
}

// KeyBindings returns the bounded keys.
func (v *Table) KeyBindings() KeyActions {
	return v.actions
}

// SearchBuff returns the associated command buffer.
func (v *Table) SearchBuff() *CmdBuff {
	return v.cmdBuff
}

// ClearSelection reset selected row.
func (v *Table) ClearSelection() {
	v.Select(0, 0)
	v.ScrollToBeginning()
}

// SelectFirstRow select first data row if any.
func (v *Table) SelectFirstRow() {
	if v.GetRowCount() > 0 {
		v.Select(1, 0)
	}
}

// ShowDeleted marks row as deleted.
func (v *Table) ShowDeleted() {
	r, _ := v.GetSelection()
	cols := v.GetColumnCount()
	for x := 0; x < cols; x++ {
		v.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// SetActions sets up keyboard action listener.
func (v *Table) SetActions(aa KeyActions) {
	for k, a := range aa {
		v.actions[k] = a
	}
}

// RmAction delete a keyed action.
func (v *Table) RmAction(kk ...tcell.Key) {
	for _, k := range kk {
		delete(v.actions, k)
	}
}

// Hints options
func (v *Table) Hints() Hints {
	if v.actions != nil {
		return v.actions.Hints()
	}

	return nil
}

// UpdateTitle refreshes the table title.
func (v *Table) UpdateTitle() {
	var title string

	rc := v.GetRowCount()
	if rc > 0 {
		rc--
	}
	switch v.activeNS {
	case resource.NotNamespaced, "*":
		title = skinTitle(fmt.Sprintf(titleFmt, v.baseTitle, rc), v.styles.Frame())
	default:
		ns := v.activeNS
		if ns == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = skinTitle(fmt.Sprintf(nsTitleFmt, v.baseTitle, ns, rc), v.styles.Frame())
	}

	if !v.cmdBuff.Empty() {
		cmd := v.cmdBuff.String()
		if isLabelSelector(cmd) {
			cmd = trimLabelSelector(cmd)
		}
		title += skinTitle(fmt.Sprintf(searchFmt, cmd), v.styles.Frame())
	}
	v.SetTitle(title)
}

// SortInvertCmd reverses sorting order.
func (v *Table) SortInvertCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.asc = !v.sortCol.asc
	v.Refresh()

	return nil
}
