package views

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	titleFmt   = " [aqua::b]%s[aqua::-][[fuchsia::b]%d[aqua::-]] "
	searchFmt  = "<[green::b]/%s[aqua::]> "
	nsTitleFmt = " [aqua::b]%s([fuchsia::b]%s[aqua::-])[aqua::-][[aqua::b]%d[aqua::-]][aqua::-] "
)

type (
	sortFn    func(rows resource.Rows, sortCol sortColumn)
	cleanseFn func(string) string

	sortColumn struct {
		index    int
		colCount int
		asc      bool
	}

	tableView struct {
		*tview.Table

		app       *appView
		baseTitle string
		currentNS string
		refreshMX sync.Mutex
		actions   keyActions
		colorerFn colorerFn
		sortFn    sortFn
		cleanseFn cleanseFn
		data      resource.TableData
		cmdBuff   *cmdBuff
		sortBuff  *cmdBuff
		tableMX   sync.Mutex
		sortCol   sortColumn
	}
)

func newTableView(app *appView, title string) *tableView {
	v := tableView{app: app, Table: tview.NewTable(), sortCol: sortColumn{0, 0, true}}
	{
		v.baseTitle = title
		v.actions = make(keyActions)
		v.SetBorder(true)
		v.SetFixed(1, 0)
		v.SetBorderColor(tcell.ColorDodgerBlue)
		v.SetBorderAttributes(tcell.AttrBold)
		v.SetBorderPadding(0, 0, 1, 1)
		v.cmdBuff = newCmdBuff('/')
		v.cmdBuff.addListener(app.cmdView)
		v.cmdBuff.reset()
		v.SetSelectable(true, false)
		v.SetSelectedStyle(tcell.ColorBlack, tcell.ColorAqua, tcell.AttrBold)
		v.SetInputCapture(v.keyboard)
		v.bindKeys()
	}

	return &v
}

func (v *tableView) bindKeys() {
	v.actions[KeyShiftI] = newKeyAction("Invert", v.sortInvertCmd, true)
	v.actions[KeyShiftN] = newKeyAction("Sort Name", v.sortColCmd(0), true)
	v.actions[KeyShiftA] = newKeyAction("Sort Age", v.sortColCmd(-1), true)

	v.actions[KeySlash] = newKeyAction("Filter Mode", v.activateCmd, false)
	v.actions[tcell.KeyEscape] = newKeyAction("Filter Reset", v.resetCmd, false)
	v.actions[tcell.KeyEnter] = newKeyAction("Filter", v.filterCmd, false)

	v.actions[tcell.KeyBackspace2] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[tcell.KeyBackspace] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[tcell.KeyDelete] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[KeyG] = newKeyAction("Top", v.app.puntCmd, false)
	v.actions[KeyShiftG] = newKeyAction("Bottom", v.app.puntCmd, false)
	v.actions[KeyB] = newKeyAction("Down", v.pageDownCmd, false)
	v.actions[KeyF] = newKeyAction("Up", v.pageUpCmd, false)
}

func (v *tableView) clearSelection() {
	v.Select(0, 0)
	v.ScrollToBeginning()
}

func (v *tableView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if key == tcell.KeyRune {
		if v.cmdBuff.isActive() {
			v.cmdBuff.add(evt.Rune())
			v.clearSelection()
			v.doUpdate(v.filtered())
			v.setSelection()
			return nil
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		log.Debug().Msgf(">> TableView handled %s", tcell.KeyNames[key])
		return a.action(evt)
	}

	return evt
}

func (v *tableView) setSelection() {
	if v.GetRowCount() > 0 {
		v.Select(1, 0)
	}
}

func (v *tableView) pageUpCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.PageUp()

	return nil
}

func (v *tableView) pageDownCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.PageDown()

	return nil
}

func (v *tableView) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.cmdBuff.setActive(false)
	v.refresh()

	return nil
}

func (v *tableView) eraseCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.isActive() {
		v.cmdBuff.del()
	}

	return nil
}

func (v *tableView) resetCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.cmdBuff.empty() {
		v.app.flash(flashInfo, "Clearing filter...")
	}
	v.cmdBuff.reset()
	v.refresh()

	return nil
}

func (v *tableView) nameColIndex() int {
	col := 0
	if v.currentNS == resource.AllNamespaces {
		col++
	}
	return col
}

func (v *tableView) sortColCmd(col int) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		if col == -1 {
			v.sortCol.index, v.sortCol.asc = v.GetColumnCount()-1, true
		} else {
			v.sortCol.index, v.sortCol.asc = v.nameColIndex()+col, true
		}

		v.refresh()
		return nil
	}
}

func (v *tableView) sortNamespaceCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.index, v.sortCol.asc = 0, true
	v.refresh()

	return nil
}

func (v *tableView) sortInvertCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.asc = !v.sortCol.asc
	v.refresh()

	return nil
}

func (v *tableView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.cmdView.inCmdMode() {
		return evt
	}

	v.app.flash(flashInfo, "Filtering...")
	v.cmdBuff.reset()
	v.cmdBuff.setActive(true)

	return nil
}

func (v *tableView) setDeleted() {
	r, _ := v.GetSelection()
	cols := v.GetColumnCount()
	for x := 0; x < cols; x++ {
		v.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// SetColorer sets up table row color management.
func (v *tableView) setColorer(f colorerFn) {
	v.colorerFn = f
}

// SetActions sets up keyboard action listener.
func (v *tableView) setActions(aa keyActions) {
	v.tableMX.Lock()
	{
		for k, a := range aa {
			v.actions[k] = a
		}
	}
	v.tableMX.Unlock()
}

// Hints options
func (v *tableView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}

	return nil
}

func (v *tableView) refresh() {
	v.update(v.data)
}

// Update table content
func (v *tableView) update(data resource.TableData) {
	v.refreshMX.Lock()
	{
		v.data = data
		if !v.cmdBuff.empty() {
			v.doUpdate(v.filtered())
		} else {
			v.doUpdate(data)
		}
	}
	v.refreshMX.Unlock()
	v.resetTitle()
}

func (v *tableView) filtered() resource.TableData {
	if v.cmdBuff.empty() {
		return v.data
	}

	rx, err := regexp.Compile(`(?i)` + v.cmdBuff.String())
	if err != nil {
		v.app.flash(flashErr, "Invalid filter expression")
		v.cmdBuff.clear()
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

func (v *tableView) sortIndicator(index int, name string) string {
	if v.sortCol.index != index {
		return name
	}

	order := "↓"
	if v.sortCol.asc {
		order = "↑"
	}
	return fmt.Sprintf("%s [green::]%s[::]", name, order)
}

func (v *tableView) doUpdate(data resource.TableData) {
	v.currentNS = data.Namespace
	if v.currentNS == resource.AllNamespaces || v.currentNS == "*" {
		v.actions[KeyShiftP] = newKeyAction("Sort Namespace", v.sortNamespaceCmd, true)
	} else {
		delete(v.actions, KeyShiftS)
	}
	v.Clear()

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

	pads := make(maxyPad, len(data.Header))
	computeMaxColumns(pads, v.sortCol.index, data)
	var row int
	for col, h := range data.Header {
		v.addHeaderCell(col, h, pads)
	}
	row++

	sortFn := v.defaultSort
	if v.sortFn != nil {
		sortFn = v.sortFn
	}
	prim, sec := v.sortAllRows(data.Rows, sortFn)
	for _, pk := range prim {
		for _, sk := range sec[pk] {
			fgColor := tcell.ColorGray
			if v.colorerFn != nil {
				fgColor = v.colorerFn(data.Namespace, data.Rows[sk])
			}
			for col, field := range data.Rows[sk].Fields {
				v.addBodyCell(row, col, field, data.Rows[sk].Deltas[col], fgColor, pads)
			}
			row++
		}
	}
}

func (v *tableView) sortAllRows(rows resource.RowEvents, sortFn sortFn) (resource.Row, map[string]resource.Row) {
	keys := make([]string, len(rows))
	v.sortRows(rows, sortFn, v.sortCol, keys)

	sec := make(map[string]resource.Row, len(rows))
	for _, k := range keys {
		grp := rows[k].Fields[v.sortCol.index]
		sec[grp] = append(sec[grp], k)
	}

	// Performs secondary to sort by name for each groups.
	prim := make(resource.Row, 0, len(sec))
	for k, v := range sec {
		sort.Strings(v)
		prim = append(prim, k)
	}
	sort.Sort(groupSorter{prim, v.sortCol.asc})

	return prim, sec
}

func (v *tableView) addHeaderCell(col int, name string, pads maxyPad) {
	c := tview.NewTableCell(v.sortIndicator(col, name))
	{
		c.SetExpansion(1)
		c.SetTextColor(tcell.ColorAntiqueWhite)
	}
	v.SetCell(0, col, c)
}

func (v *tableView) addBodyCell(row, col int, field, delta string, color tcell.Color, pads maxyPad) {
	var pField string
	if isASCII(field) {
		pField = pad(deltas(delta, field), pads[col])
	} else {
		pField = deltas(delta, field)
	}

	c := tview.NewTableCell(pField)
	{
		c.SetExpansion(1)
		c.SetTextColor(color)
	}
	v.SetCell(row, col, c)
}

func (v *tableView) defaultSort(rows resource.Rows, sortCol sortColumn) {
	t := rowSorter{rows: rows, index: sortCol.index, asc: sortCol.asc}
	sort.Sort(t)
}

func (*tableView) sortRows(evts resource.RowEvents, sortFn sortFn, sortCol sortColumn, keys []string) {
	rows := make(resource.Rows, 0, len(evts))
	for k, r := range evts {
		rows = append(rows, append(r.Fields, k))
	}
	sortFn(rows, sortCol)

	for i, r := range rows {
		keys[i] = r[len(r)-1]
	}
}

func (*tableView) defaultColCleanse(s string) string {
	return strings.TrimSpace(s)
}

func (v *tableView) resetTitle() {
	var title string

	rc := v.GetRowCount()
	if rc > 0 {
		rc--
	}
	switch v.currentNS {
	case resource.NotNamespaced:
		title = fmt.Sprintf(titleFmt, v.baseTitle, rc)
	default:
		ns := v.currentNS
		if v.currentNS == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = fmt.Sprintf(nsTitleFmt, v.baseTitle, ns, rc)
	}

	if !v.cmdBuff.isActive() && !v.cmdBuff.empty() {
		title += fmt.Sprintf(searchFmt, v.cmdBuff)
	}
	v.SetTitle(title)
}

// ----------------------------------------------------------------------------
// Event listeners...

func (v *tableView) changed(s string) {}

func (v *tableView) active(b bool) {
	if b {
		v.SetBorderColor(tcell.ColorRed)
		return
	}
	v.SetBorderColor(tcell.ColorDodgerBlue)
}
