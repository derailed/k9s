package views

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/util/duration"
)

type (
	resTable struct {
		*tview.Table

		app       *appView
		baseTitle string
		currentNS string
		data      resource.TableData
		actions   keyActions
	}

	tableView struct {
		*resTable

		cmdBuff   *cmdBuff
		colorerFn colorerFn
		sortFn    sortFn
		cleanseFn cleanseFn
		filterFn  func(string)
		sortCol   sortColumn
	}
)

func newTableView(app *appView, title string) *tableView {
	v := tableView{
		resTable: &resTable{
			Table:     tview.NewTable(),
			app:       app,
			actions:   make(keyActions),
			baseTitle: title,
		},
		sortCol: sortColumn{0, 0, true},
		cmdBuff: newCmdBuff('/'),
	}
	v.SetFixed(1, 0)
	v.SetBorder(true)
	v.SetBackgroundColor(config.AsColor(app.styles.Table().BgColor))
	v.SetBorderColor(config.AsColor(app.styles.Table().FgColor))
	v.SetBorderFocusColor(config.AsColor(app.styles.Frame().Border.FocusColor))
	v.SetBorderAttributes(tcell.AttrBold)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetSelectable(true, false)
	v.SetSelectedStyle(
		tcell.ColorBlack,
		config.AsColor(app.styles.Table().CursorColor),
		tcell.AttrBold,
	)
	v.cmdBuff.addListener(app.cmd())
	v.cmdBuff.reset()

	v.SetInputCapture(v.keyboard)
	v.bindKeys()

	return &v
}

func (v *tableView) bindKeys() {
	v.actions = keyActions{
		tcell.KeyCtrlS:      newKeyAction("Save", v.saveCmd, true),
		KeySlash:            newKeyAction("Filter Mode", v.activateCmd, false),
		tcell.KeyEscape:     newKeyAction("Filter Reset", v.resetCmd, false),
		tcell.KeyEnter:      newKeyAction("Filter", v.filterCmd, false),
		tcell.KeyBackspace2: newKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyBackspace:  newKeyAction("Erase", v.eraseCmd, false),
		tcell.KeyDelete:     newKeyAction("Erase", v.eraseCmd, false),
		KeyShiftI:           newKeyAction("Invert", v.sortInvertCmd, false),
		KeyShiftN:           newKeyAction("Sort Name", v.sortColCmd(0), true),
		KeyShiftA:           newKeyAction("Sort Age", v.sortColCmd(-1), true),
	}
}

func (v *tableView) setFilterFn(fn func(string)) {
	v.filterFn = fn
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
			v.selectFirstRow()
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

func (v *tableView) selectFirstRow() {
	if v.GetRowCount() > 0 {
		v.Select(1, 0)
	}
}

func (v *tableView) saveCmd(evt *tcell.EventKey) *tcell.EventKey {
	if path, err := saveTable(v.app.config.K9s.CurrentCluster, v.baseTitle, v.data); err != nil {
		v.app.flash().err(err)
	} else {
		v.app.flash().infof("File %s saved successfully!", path)
	}

	return nil
}

func (v *tableView) filterCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.cmdBuff.isActive() {
		v.cmdBuff.setActive(false)
		cmd := v.cmdBuff.String()
		if isLabelSelector(cmd) && v.filterFn != nil {
			v.filterFn(trimLabelSelector(cmd))
			return nil
		}
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
		v.app.flash().info("Clearing filter...")
	}
	if isLabelSelector(v.cmdBuff.String()) {
		v.filterFn("")
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

func (v *tableView) sortInvertCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.sortCol.asc = !v.sortCol.asc
	v.refresh()

	return nil
}

func (v *tableView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.inCmdMode() {
		return evt
	}

	v.app.flash().info("Filter mode activated.")
	if isLabelSelector(v.cmdBuff.String()) {
		return nil
	}
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
	if v.cmdBuff.empty() || isLabelSelector(v.cmdBuff.String()) {
		return v.data
	}

	rx, err := regexp.Compile(`(?i)` + v.cmdBuff.String())
	if err != nil {
		v.app.flash().err(errors.New("Invalid filter expression"))
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

func (v *tableView) adjustSorter(data resource.TableData) {
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

func (v *tableView) doUpdate(data resource.TableData) {
	v.currentNS = data.Namespace
	if v.currentNS == resource.AllNamespaces && v.currentNS != "*" {
		v.actions[KeyShiftP] = newKeyAction("Sort Namespace", v.sortColCmd(0), true)
	} else {
		delete(v.actions, KeyShiftP)
	}
	v.Clear()

	v.adjustSorter(data)

	var row int
	fg := config.AsColor(v.app.styles.Table().Header.FgColor)
	bg := config.AsColor(v.app.styles.Table().Header.BgColor)
	for col, h := range data.Header {
		v.addHeaderCell(data.NumCols[h], col, h)
		c := v.GetCell(0, col)
		c.SetBackgroundColor(bg)
		c.SetTextColor(fg)
	}
	row++

	v.sort(data, row)
}

func (v *tableView) sort(data resource.TableData, row int) {
	pads := make(maxyPad, len(data.Header))
	computeMaxColumns(pads, v.sortCol.index, data)

	sortFn := defaultSort
	if v.sortFn != nil {
		sortFn = v.sortFn
	}
	prim, sec := sortAllRows(v.sortCol, data.Rows, sortFn)
	fgColor := config.AsColor(v.app.styles.Table().FgColor)
	for _, pk := range prim {
		for _, sk := range sec[pk] {
			if v.colorerFn != nil {
				fgColor = v.colorerFn(data.Namespace, data.Rows[sk])
			}
			for col, field := range data.Rows[sk].Fields {
				header := data.Header[col]
				field, align := v.formatCell(data.NumCols[header], header, field+deltas(data.Rows[sk].Deltas[col], field), pads[col])
				c := tview.NewTableCell(field)
				{
					c.SetExpansion(1)
					c.SetAlign(align)
					c.SetTextColor(fgColor)
				}
				v.SetCell(row, col, c)
			}
			row++
		}
	}
}

func (v *tableView) addHeaderCell(numerical bool, col int, name string) {
	c := tview.NewTableCell(sortIndicator(v.sortCol, v.app.styles.Table(), col, name))
	c.SetExpansion(1)
	if numerical || cpuRX.MatchString(name) || memRX.MatchString(name) {
		c.SetAlign(tview.AlignRight)
	}
	v.SetCell(0, col, c)
}

func (v *tableView) formatCell(numerical bool, header, field string, padding int) (string, int) {
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
	if isASCII(field) {
		return pad(field, padding), align
	}

	return field, align
}

func (v *tableView) resetTitle() {
	var title string

	rc := v.GetRowCount()
	if rc > 0 {
		rc--
	}
	switch v.currentNS {
	case resource.NotNamespaced, "*":
		title = skinTitle(fmt.Sprintf(titleFmt, v.baseTitle, rc), v.app.styles.Frame())
	default:
		ns := v.currentNS
		if ns == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = skinTitle(fmt.Sprintf(nsTitleFmt, v.baseTitle, ns, rc), v.app.styles.Frame())
	}

	if !v.cmdBuff.isActive() && !v.cmdBuff.empty() {
		cmd := v.cmdBuff.String()
		if isLabelSelector(cmd) {
			cmd = trimLabelSelector(cmd)
		}
		title += skinTitle(fmt.Sprintf(searchFmt, cmd), v.app.styles.Frame())
	}
	v.SetTitle(title)
}
