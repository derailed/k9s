package views

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/duration"
)

const (
	titleFmt      = "[fg:bg:b] %s[fg:bg:-][[count:bg:b]%d[fg:bg:-]] "
	searchFmt     = "<[filter:bg:b]/%s[fg:bg:]> "
	nsTitleFmt    = "[fg:bg:b] %s([hilite:bg:b]%s[fg:bg:-])[fg:bg:-][[count:bg:b]%d[fg:bg:-]][fg:bg:-] "
	descIndicator = "↓"
	ascIndicator  = "↑"
)

var (
	crx = regexp.MustCompile(`\A.{0,1}CPU`)
	mrx = regexp.MustCompile(`\A.{0,1}MEM`)
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
		actions   keyActions
		colorerFn colorerFn
		sortFn    sortFn
		cleanseFn cleanseFn
		data      resource.TableData
		cmdBuff   *cmdBuff
		sortBuff  *cmdBuff
		sortCol   sortColumn
	}
)

func newTableView(app *appView, title string) *tableView {
	v := tableView{app: app, Table: tview.NewTable(), sortCol: sortColumn{0, 0, true}}
	{
		v.baseTitle = title
		v.actions = make(keyActions)
		v.SetFixed(1, 0)
		v.SetBorder(true)
		v.SetBackgroundColor(config.AsColor(app.styles.Style.Table.BgColor))
		v.SetBorderColor(config.AsColor(app.styles.Style.Table.FgColor))
		v.SetBorderFocusColor(config.AsColor(app.styles.Style.Border.FocusColor))
		v.SetBorderAttributes(tcell.AttrBold)
		v.SetBorderPadding(0, 0, 1, 1)
		v.cmdBuff = newCmdBuff('/')
		v.cmdBuff.addListener(app.cmdView)
		v.cmdBuff.reset()
		v.SetSelectable(true, false)
		v.SetSelectedStyle(
			tcell.ColorBlack,
			config.AsColor(app.styles.Style.Table.CursorColor),
			tcell.AttrBold,
		)
		v.SetInputCapture(v.keyboard)
		v.bindKeys()
	}

	return &v
}

func (v *tableView) bindKeys() {
	v.actions[KeyShiftI] = newKeyAction("Invert", v.sortInvertCmd, false)
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
		if evt.Modifiers() == tcell.ModAlt {
			key = tcell.Key(int16(evt.Rune()) * int16(evt.Modifiers()))
		}
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
	if v.cmdBuff.isActive() {
		v.cmdBuff.setActive(false)
		v.refresh()
		return nil
	}

	return evt
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

	v.app.flash(flashInfo, "Filter mode activated.")
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
	for k, a := range aa {
		v.actions[k] = a
	}
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
	v.data = data
	if !v.cmdBuff.empty() {
		v.doUpdate(v.filtered())
	} else {
		v.doUpdate(v.data)
	}
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

	order := descIndicator
	if v.sortCol.asc {
		order = ascIndicator
	}
	return fmt.Sprintf("%s[%s::]%s[::]", name, v.app.styles.Style.Table.Header.SorterColor, order)
}

func (v *tableView) doUpdate(data resource.TableData) {
	v.currentNS = data.Namespace
	if v.currentNS == resource.AllNamespaces && v.currentNS != "*" {
		v.actions[KeyShiftP] = newKeyAction("Sort Namespace", v.sortNamespaceCmd, true)
	} else {
		delete(v.actions, KeyShiftP)
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
	fg := config.AsColor(v.app.styles.Style.Table.Header.FgColor)
	bg := config.AsColor(v.app.styles.Style.Table.Header.BgColor)
	for col, h := range data.Header {
		v.addHeaderCell(col, h, fg, bg)
	}
	row++

	sortFn := v.defaultSort
	if v.sortFn != nil {
		sortFn = v.sortFn
	}
	prim, sec := v.sortAllRows(data.Rows, sortFn)
	fgColor := config.AsColor(v.app.styles.Style.Table.FgColor)
	for _, pk := range prim {
		for _, sk := range sec[pk] {
			if v.colorerFn != nil {
				fgColor = v.colorerFn(data.Namespace, data.Rows[sk])
			}
			for col, field := range data.Rows[sk].Fields {
				v.addBodyCell(data.Header[col], row, col, field, data.Rows[sk].Deltas[col], fgColor, pads)
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

func (v *tableView) addHeaderCell(col int, name string, fg, bg tcell.Color) {
	c := tview.NewTableCell(v.sortIndicator(col, name))
	{
		c.SetExpansion(1)
		c.SetTextColor(fg)
		if crx.MatchString(name) || mrx.MatchString(name) {
			c.SetAlign(tview.AlignRight)
		}
		c.SetBackgroundColor(bg)
	}
	v.SetCell(0, col, c)
}

func (v *tableView) addBodyCell(header string, row, col int, field, delta string, color tcell.Color, pads maxyPad) {
	const colPadding = 3

	if header == "AGE" {
		dur, err := time.ParseDuration(field)
		if err == nil {
			field = duration.HumanDuration(dur)
		}
	}

	field += deltas(delta, field)
	align := tview.AlignLeft
	if crx.MatchString(header) || mrx.MatchString(header) {
		align = tview.AlignRight
	} else if isASCII(field) {
		field = pad(field, pads[col]+colPadding)
	}

	c := tview.NewTableCell(field)
	{
		c.SetExpansion(1)
		c.SetAlign(align)
		c.SetTextColor(color)
		c.SetMaxWidth(pads[col] + colPadding)
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
	case resource.NotNamespaced, "*":
		fmat := strings.Replace(titleFmt, "[fg:bg", "["+v.app.styles.Style.Title.FgColor+":"+v.app.styles.Style.Title.BgColor, -1)
		fmat = strings.Replace(fmat, "[count", "["+v.app.styles.Style.Title.CounterColor, 1)
		title = fmt.Sprintf(fmat, v.baseTitle, rc)
	default:
		ns := v.currentNS
		if ns == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		fmat := strings.Replace(nsTitleFmt, "[fg:bg", "["+v.app.styles.Style.Title.FgColor+":"+v.app.styles.Style.Title.BgColor, -1)
		fmat = strings.Replace(fmat, "[hilite", "["+v.app.styles.Style.Title.HighlightColor, 1)
		fmat = strings.Replace(fmat, "[count", "["+v.app.styles.Style.Title.CounterColor, 1)
		title = fmt.Sprintf(fmat, v.baseTitle, ns, rc)
	}

	if !v.cmdBuff.isActive() && !v.cmdBuff.empty() {
		fmat := strings.Replace(searchFmt, "[fg:bg", "["+v.app.styles.Style.Title.FgColor+":"+v.app.styles.Style.Title.BgColor, -1)
		fmat = strings.Replace(fmat, ":bg:", ":"+v.app.styles.Style.Title.BgColor+":", -1)
		fmat = strings.Replace(fmat, "[filter", "["+v.app.styles.Style.Title.FilterColor, 1)
		title += fmt.Sprintf(fmat, v.cmdBuff)
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
