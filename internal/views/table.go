package views

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
	log "github.com/sirupsen/logrus"
)

const (
	titleFmt   = " [aqua::b]%s[aqua::-]([fuchsia::b]%d[aqua::-]) "
	nsTitleFmt = " [aqua::b]%s([fuchsia::b]%s[aqua::-])[aqua::-][[aqua::b]%d[aqua::-]][aqua::-] "
)

type (
	tableView struct {
		*tview.Flex

		app       *appView
		baseTitle string
		currentNS string
		refresh   sync.Mutex
		actions   keyActions
		colorer   colorerFn
		sortFn    resource.SortFn
		table     *tview.Table
		data      resource.TableData
		cmdBuff   *cmdBuff
	}
)

func newTableView(app *appView, title string, sortFn resource.SortFn) *tableView {
	v := tableView{app: app, Flex: tview.NewFlex().SetDirection(tview.FlexRow)}
	{
		v.baseTitle = title
		v.sortFn = sortFn
		v.SetBorder(true)
		v.SetBorderColor(tcell.ColorDodgerBlue)
		v.SetBorderAttributes(tcell.AttrBold)
		v.SetBorderPadding(0, 0, 1, 1)
		v.cmdBuff = newCmdBuff('/')
	}

	v.cmdBuff.addListener(app.cmdView)
	v.cmdBuff.reset()

	v.table = tview.NewTable()
	{
		v.table.SetSelectable(true, false)
		v.table.SetSelectedStyle(tcell.ColorBlack, tcell.ColorAqua, tcell.AttrBold)
		v.table.SetInputCapture(v.keyboard)
	}

	v.AddItem(v.table, 0, 1, true)
	return &v
}

func (v *tableView) setDeleted() {
	r, _ := v.table.GetSelection()
	cols := v.table.GetColumnCount()
	for x := 0; x < cols; x++ {
		v.table.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

func (v *tableView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if evt.Key() == tcell.KeyRune {
		if v.cmdBuff.isActive() {
			v.cmdBuff.add(evt.Rune())
		}
		switch evt.Rune() {
		case v.cmdBuff.hotKey:
			if !v.app.cmdView.inCmdMode() {
				v.app.flash(flashInfo, "Entering filtering mode...")
				log.Info("K9s entering filtering mode...")
				v.cmdBuff.setActive(true)
			}
			return evt
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		if !v.app.cmdView.inCmdMode() {
			a.action(evt)
		}
	}

	switch evt.Key() {
	case tcell.KeyEnter:
		if v.cmdBuff.isActive() && !v.cmdBuff.empty() {
			v.filter()
		}
		v.cmdBuff.setActive(false)
	case tcell.KeyEsc:
		v.cmdBuff.reset()
		v.filter()
	case tcell.KeyBackspace2:
		if v.cmdBuff.isActive() {
			v.cmdBuff.del()
		}
	}
	return evt
}

func (v *tableView) filter() {
	v.filterData(v.cmdBuff)
}

func (v *tableView) filterData(filter fmt.Stringer) {
	filtered := resource.TableData{
		Header:    v.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: v.data.Namespace,
	}

	rx, err := regexp.Compile(filter.String())
	if err != nil {
		v.app.flash(flashErr, "Invalid search expression")
		v.cmdBuff.clear()
		return
	}
	for k, row := range v.data.Rows {
		f := strings.Join(row.Fields, " ")
		if rx.MatchString(f) {
			filtered.Rows[k] = row
		}
	}
	v.doUpdate(filtered)
}

// SetColorer sets up table row color management.
func (v *tableView) SetColorer(f colorerFn) {
	v.colorer = f
}

// AddActions sets up keyboard action listener.
func (v *tableView) addActions(kk keyActions) {
	for k, a := range kk {
		v.actions[k] = a
	}
}

// SetActions sets up keyboard action listener.
func (v *tableView) setActions(aa keyActions) {
	v.actions = aa
}

// Hints options
func (v *tableView) hints() hints {
	if v.actions != nil {
		return v.actions.toHints()
	}
	return nil
}

func (v *tableView) resetTitle() {
	var title string

	switch v.currentNS {
	case resource.NotNamespaced:
		title = fmt.Sprintf(titleFmt, v.baseTitle, v.table.GetRowCount()-1)
	default:
		ns := v.currentNS
		if v.currentNS == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = fmt.Sprintf(nsTitleFmt, v.baseTitle, ns, v.table.GetRowCount()-1)
	}

	if !v.cmdBuff.empty() {
		title += fmt.Sprintf("<[green::b]/%s[aqua::]> ", v.cmdBuff)
	}
	v.SetTitle(title)
}

// Update table content
func (v *tableView) update(data resource.TableData) {
	v.refresh.Lock()
	{
		v.data = data
		if !v.cmdBuff.empty() {
			v.filter()
		} else {
			v.doUpdate(data)
		}
		v.resetTitle()
	}
	v.refresh.Unlock()
}

func (v *tableView) doUpdate(data resource.TableData) {
	v.table.Clear()
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
		v.table.SetCell(row, col, c)
	}
	row++

	keys := make([]string, 0, len(data.Rows))
	for k := range data.Rows {
		keys = append(keys, k)
	}
	v.sortFn(keys)
	for _, k := range keys {
		fgColor := tcell.ColorGray
		if v.colorer != nil {
			fgColor = v.colorer(data.Namespace, data.Rows[k])
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
			v.table.SetCell(row, col, c)
		}
		row++
	}
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
