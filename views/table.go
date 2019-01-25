package views

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell"
	"github.com/k8sland/k9s/resource"
	"github.com/k8sland/tview"
)

const (
	titleFmt   = " [aqua::b]%s[aqua::-]([fuchsia::b]%d[aqua::-]) "
	nsTitleFmt = " [aqua::-]<[fuchsia::b]%s[aqua::-]>" + titleFmt
)

type (
	tableView struct {
		*tview.Table
		baseTitle string
		currentNS string
		actions   keyActions
		colorer   colorerFn
		cmdBuff   string
		sortFn    resource.SortFn
		parent    *resourceView
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
	v.Select(0, 0)
}

func (v *tableView) keyboard(evt *tcell.EventKey) *tcell.EventKey {
	if evt.Key() == tcell.KeyRune {
		if a, ok := v.actions[tcell.Key(evt.Rune())]; ok {
			a.action(evt)
			evt = nil
		}
		return evt
	}

	if a, ok := v.actions[evt.Key()]; ok {
		a.action(evt)
		evt = nil
	}
	return evt
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
	switch v.currentNS {
	case resource.NotNamespaced:
		v.SetTitle(fmt.Sprintf(titleFmt, v.baseTitle, v.GetRowCount()-1))
	default:
		ns := v.currentNS
		if v.currentNS == resource.AllNamespaces {
			ns = "all"
		}
		v.SetTitle(fmt.Sprintf(nsTitleFmt, strings.Title(ns), v.baseTitle, v.GetRowCount()-1))
	}
}

// Update table content
func (v *tableView) update(data resource.TableData) {
	v.Clear()
	v.currentNS = data.Namespace

	var row int
	for col, h := range data.Header {
		c := tview.NewTableCell(h)
		if len(h) == 0 {
			c.SetExpansion(1)
		} else {
			c.SetExpansion(3)
		}
		c.SetTextColor(tcell.ColorWhite)
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
			if len(data.Header[col]) == 0 {
				c.SetExpansion(1)
			} else {
				c.SetExpansion(3)
			}
			c.SetTextColor(fgColor)
			v.SetCell(row, col, c)
		}
		row++
	}
	v.resetTitle()
}
