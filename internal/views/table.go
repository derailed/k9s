package views

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

const (
	titleFmt   = " [aqua::b]%s[aqua::-]([fuchsia::b]%d[aqua::-]) "
	searchFmt  = "<[green::b]/%s[aqua::]> "
	nsTitleFmt = " [aqua::b]%s([fuchsia::b]%s[aqua::-])[aqua::-][[aqua::b]%d[aqua::-]][aqua::-] "
)

type (
	tableView struct {
		*tview.Table

		app       *appView
		baseTitle string
		currentNS string
		refreshMX sync.Mutex
		actions   keyActions
		colorerFn colorerFn
		sortFn    resource.SortFn
		data      resource.TableData
		cmdBuff   *cmdBuff
		tableMX   sync.Mutex
	}
)

func newTableView(app *appView, title string, sortFn resource.SortFn) *tableView {
	v := tableView{app: app, Table: tview.NewTable()}
	{
		v.baseTitle = title
		v.sortFn = sortFn
		v.actions = make(keyActions)
		v.SetBorder(true)
		v.SetBorderColor(tcell.ColorDodgerBlue)
		v.SetBorderAttributes(tcell.AttrBold)
		v.SetBorderPadding(0, 0, 1, 1)
		v.cmdBuff = newCmdBuff('/')
		v.cmdBuff.addListener(app.cmdView)
		v.cmdBuff.reset()
		v.SetSelectable(true, false)
		v.SetSelectedStyle(tcell.ColorBlack, tcell.ColorAqua, tcell.AttrBold)
		v.SetInputCapture(v.keyboard)
	}

	v.actions[KeySlash] = newKeyAction("Filter", v.activateCmd, false)
	v.actions[tcell.KeyEnter] = newKeyAction("Search", v.filterCmd, false)
	v.actions[tcell.KeyEscape] = newKeyAction("Reset Filter", v.resetCmd, false)
	v.actions[tcell.KeyBackspace2] = newKeyAction("Erase", v.eraseCmd, false)
	v.actions[KeyG] = newKeyAction("Top", app.puntCmd, false)
	v.actions[KeyShiftG] = newKeyAction("Bottom", app.puntCmd, false)
	v.actions[KeyB] = newKeyAction("Down", v.pageDownCmd, false)
	v.actions[KeyF] = newKeyAction("Up", v.pageUpCmd, false)

	return &v
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

func (v *tableView) activateCmd(evt *tcell.EventKey) *tcell.EventKey {
	if v.app.cmdView.inCmdMode() {
		return evt
	}

	v.app.flash(flashInfo, "Filtering...")
	log.Info().Msg("Entering filtering mode...")
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
func (v *tableView) SetColorer(f colorerFn) {
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

func (v *tableView) doUpdate(data resource.TableData) {
	v.Clear()
	v.currentNS = data.Namespace

	var row int
	for col, h := range data.Header {
		c := tview.NewTableCell(h)
		{
			c.SetExpansion(3)
			if len(h) == 0 {
				c.SetExpansion(1)
			}
			c.SetTextColor(tcell.ColorWhite)
		}
		v.SetCell(row, col, c)
	}
	row++

	keys := make([]string, 0, len(data.Rows))
	for k := range data.Rows {
		keys = append(keys, k)
	}
	if v.sortFn != nil {
		v.sortFn(keys)
	}
	for _, k := range keys {
		fgColor := tcell.ColorGray
		if v.colorerFn != nil {
			fgColor = v.colorerFn(data.Namespace, data.Rows[k])
		}
		for col, f := range data.Rows[k].Fields {
			c := tview.NewTableCell(deltas(data.Rows[k].Deltas[col], f))
			{
				c.SetExpansion(3)
				if len(data.Header[col]) == 0 {
					c.SetExpansion(1)
				}
				c.SetTextColor(fgColor)
			}
			v.SetCell(row, col, c)
		}
		row++
	}
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

func (v *tableView) changed(s string) {
}

func (v *tableView) active(b bool) {
	if b {
		v.SetBorderColor(tcell.ColorRed)
		return
	}
	v.SetBorderColor(tcell.ColorDodgerBlue)
}
