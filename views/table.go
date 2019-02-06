package views

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	"github.com/k8sland/tview"
)

const (
	titleFmt   = " [aqua::b]%s[aqua::-]([fuchsia::b]%d[aqua::-]) "
	nsTitleFmt = " [aqua::b]%s([fuchsia::b]%s[aqua::-])[aqua::-][[aqua::b]%d[aqua::-]][aqua::-] "
)

type (
	tableView struct {
		*tview.Table
		baseTitle  string
		currentNS  string
		actions    keyActions
		colorer    colorerFn
		sortFn     resource.SortFn
		parent     *resourceView
		cmdBuffer  []rune
		data       resource.TableData
		searchMode bool
		filtered   bool
	}
)

func newTableView(title string, sortFn resource.SortFn) *tableView {
	v := tableView{Table: tview.NewTable(), baseTitle: title, sortFn: sortFn}
	v.SetBorder(true)
	v.SetBorderColor(tcell.ColorDodgerBlue)
	v.SetBorderAttributes(tcell.AttrBold)
	v.SetBorderPadding(0, 0, 1, 1)
	v.SetSelectable(true, false)
	v.SetSelectedStyle(tcell.ColorBlack, tcell.ColorAqua, tcell.AttrBold)
	v.SetInputCapture(v.keyboard)
	return &v
}

func (v *tableView) setDeleted() {
	r, _ := v.GetSelection()
	cols := v.GetColumnCount()
	for x := 0; x < cols; x++ {
		v.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

func (v *tableView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	key := evt.Key()
	if evt.Key() == tcell.KeyRune {
		if evt.Rune() == '/' {
			v.searchMode = true
			v.cmdBuffer = []rune{}
		} else {
			if v.searchMode {
				v.cmdBuffer = append([]rune(v.cmdBuffer), evt.Rune())
			}
		}
		key = tcell.Key(evt.Rune())
	}

	if a, ok := v.actions[key]; ok {
		a.action(evt)
		return nil
	}

	switch evt.Key() {
	case tcell.KeyEnter:
		v.filtered = true
		v.filter()
		v.searchMode = false
		evt = nil
	case tcell.KeyEsc:
		v.filtered, v.searchMode = false, false
		v.cmdBuffer = []rune{}
		evt = nil
	}
	return evt
}

func (v *tableView) filter() {
	v.filterData(string(v.cmdBuffer))
}

func (v *tableView) filterData(filter string) {
	filtered := resource.TableData{
		Header:    v.data.Header,
		Rows:      resource.RowEvents{},
		Namespace: v.data.Namespace,
	}

	rx := regexp.MustCompile(filter)
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
		title = fmt.Sprintf(titleFmt, v.baseTitle, v.GetRowCount()-1)
	default:
		ns := v.currentNS
		if v.currentNS == resource.AllNamespaces {
			ns = resource.AllNamespace
		}
		title = fmt.Sprintf(nsTitleFmt, v.baseTitle, ns, v.GetRowCount()-1)
	}

	if v.filtered {
		title += fmt.Sprintf("<[green::b]/%s[aqua::]> ", string(v.cmdBuffer))
	}
	v.SetTitle(title)
}

// Update table content
func (v *tableView) update(data resource.TableData) {
	v.data = data
	if v.filtered {
		v.filter()
	} else {
		v.doUpdate(data)
	}
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
			v.SetCell(row, col, c)
		}
		row++
	}
	v.resetTitle()
}
