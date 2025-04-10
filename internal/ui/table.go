// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/vul"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

const maxTruncate = 50

type (
	// ColorerFunc represents a row colorer.
	ColorerFunc func(ns string, evt model1.RowEvent) tcell.Color

	// DecorateFunc represents a row decorator.
	DecorateFunc func(*model1.TableData)

	// SelectedRowFunc a table selection callback.
	SelectedRowFunc func(r int)
)

// Table represents tabular data.
type Table struct {
	*SelectTable
	gvr         *client.GVR
	sortCol     model1.SortColumn
	manualSort  bool
	Path        string
	Extras      string
	actions     *KeyActions
	cmdBuff     *model.FishBuff
	styles      *config.Styles
	viewSetting *config.ViewSetting
	colorerFn   model1.ColorerFunc
	decorateFn  DecorateFunc
	wide        bool
	toast       bool
	hasMetrics  bool
	ctx         context.Context
	mx          sync.RWMutex
	readOnly    bool
	noIcon      bool
	fullGVR     bool
}

// NewTable returns a new table view.
func NewTable(gvr *client.GVR) *Table {
	return &Table{
		SelectTable: &SelectTable{
			Table: tview.NewTable(),
			model: model.NewTable(gvr),
			marks: make(map[string]struct{}),
		},
		ctx:     context.Background(),
		gvr:     gvr,
		actions: NewKeyActions(),
		cmdBuff: model.NewFishBuff('/', model.FilterBuffer),
		sortCol: model1.SortColumn{ASC: true},
	}
}

// SetFullGVR toggles full GVR title display.
func (t *Table) SetFullGVR(b bool) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.fullGVR = b
}

// SetNoIcon toggles no icon mode.
func (t *Table) SetNoIcon(b bool) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.noIcon = b
}

// SetReadOnly toggles read-only mode.
func (t *Table) SetReadOnly(ro bool) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.readOnly = ro
}

func (t *Table) setSortCol(sc model1.SortColumn) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.sortCol = sc
}

func (t *Table) toggleSortCol() {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.sortCol.ASC = !t.sortCol.ASC
}

func (t *Table) getSortCol() model1.SortColumn {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.sortCol
}

func (t *Table) setMSort(b bool) {
	t.mx.Lock()
	defer t.mx.Unlock()

	t.manualSort = b
}

func (t *Table) getMSort() bool {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.manualSort
}

// SetViewSetting sets custom view config is present.
func (t *Table) SetViewSetting(vs *config.ViewSetting) bool {
	t.mx.Lock()
	defer t.mx.Unlock()

	if !t.viewSetting.Equals(vs) {
		t.viewSetting = vs
		slog.Debug("Updating custom view setting", slogs.GVR, t.gvr, slogs.ViewSetting, vs)
		t.model.SetViewSetting(t.ctx, vs)
		return true
	}

	return false
}

// GetViewSetting return current view settings if any.
func (t *Table) GetViewSetting() *config.ViewSetting {
	t.mx.RLock()
	defer t.mx.RUnlock()

	return t.viewSetting
}

func (t *Table) GetContext() context.Context {
	return t.ctx
}

func (t *Table) SetContext(ctx context.Context) {
	t.ctx = ctx
}

// Init initializes the component.
func (t *Table) Init(ctx context.Context) {
	t.SetFixed(1, 0)
	t.SetBorder(true)
	t.SetBorderAttributes(tcell.AttrBold)
	t.SetBorderPadding(0, 0, 1, 1)
	t.SetSelectable(true, false)
	t.SetSelectionChangedFunc(t.selectionChanged)
	t.SetBackgroundColor(tcell.ColorDefault)
	t.Select(1, 0)
	t.styles = mustExtractStyles(ctx)
	t.StylesChanged(t.styles)
}

// GVR returns a resource descriptor.
func (t *Table) GVR() *client.GVR { return t.gvr }

// ViewSettingsChanged notifies listener the view configuration changed.
func (t *Table) ViewSettingsChanged(vs *config.ViewSetting) {
	if t.SetViewSetting(vs) {
		if vs == nil {
			if !t.getMSort() && !t.sortCol.IsSet() {
				t.setSortCol(model1.SortColumn{})
			}
		} else {
			t.setMSort(false)
		}
		t.Refresh()
	}
}

// StylesChanged notifies the skin changed.
func (t *Table) StylesChanged(s *config.Styles) {
	t.SetBackgroundColor(s.Table().BgColor.Color())
	t.SetBorderColor(s.Frame().Border.FgColor.Color())
	t.SetBorderFocusColor(s.Frame().Border.FocusColor.Color())
	t.SetSelectedStyle(
		tcell.StyleDefault.Foreground(t.styles.Table().CursorFgColor.Color()).
			Background(t.styles.Table().CursorBgColor.Color()).Attributes(tcell.AttrBold))
	t.selFgColor = s.Table().CursorFgColor.Color()
	t.selBgColor = s.Table().CursorBgColor.Color()
	t.Refresh()
}

// ResetToast resets toast flag.
func (t *Table) ResetToast() {
	t.toast = false
	t.Refresh()
}

// ToggleToast toggles to show toast resources.
func (t *Table) ToggleToast() {
	t.toast = !t.toast
	t.Refresh()
}

// ToggleWide toggles wide col display.
func (t *Table) ToggleWide() {
	t.wide = !t.wide
	t.Refresh()
}

// Actions returns active menu bindings.
func (t *Table) Actions() *KeyActions {
	return t.actions
}

// Styles returns styling configurator.
func (t *Table) Styles() *config.Styles {
	return t.styles
}

// FilterInput filters user commands.
func (t *Table) FilterInput(r rune) bool {
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

// Filter filters out table data.
func (t *Table) Filter(string) {
	t.ClearSelection()
	t.doUpdate(t.filtered(t.GetModel().Peek()))
	t.UpdateTitle()
	t.SelectFirstRow()
}

// Hints returns the view hints.
func (t *Table) Hints() model.MenuHints {
	return t.actions.Hints()
}

// ExtraHints returns additional hints.
func (*Table) ExtraHints() map[string]string {
	return nil
}

// GetFilteredData fetch filtered tabular data.
func (t *Table) GetFilteredData() *model1.TableData {
	return t.filtered(t.GetModel().Peek())
}

// SetDecorateFn specifies the default row decorator.
func (t *Table) SetDecorateFn(f DecorateFunc) {
	t.decorateFn = f
}

// SetColorerFn specifies the default colorer.
func (t *Table) SetColorerFn(f model1.ColorerFunc) {
	t.colorerFn = f
}

// SetSortCol sets in sort column index and order.
func (t *Table) SetSortCol(name string, asc bool) {
	t.setSortCol(model1.SortColumn{Name: name, ASC: asc})
}

// Update table content.
func (t *Table) Update(data *model1.TableData, hasMetrics bool) *model1.TableData {
	if t.decorateFn != nil {
		t.decorateFn(data)
	}
	t.hasMetrics = hasMetrics

	return t.doUpdate(t.filtered(data))
}

func (t *Table) GetNamespace() string {
	if t.GetModel() != nil {
		return t.GetModel().GetNamespace()
	}

	return client.NamespaceAll
}

func (t *Table) doUpdate(data *model1.TableData) *model1.TableData {
	if client.IsAllNamespaces(data.GetNamespace()) {
		t.actions.Add(
			KeyShiftP,
			NewKeyAction("Sort Namespace", t.SortColCmd("NAMESPACE", true), false),
		)
	} else {
		t.actions.Delete(KeyShiftP)
	}

	t.setSortCol(data.ComputeSortCol(t.GetViewSetting(), t.getSortCol(), t.getMSort()))

	return data
}

func (t *Table) UpdateUI(cdata, data *model1.TableData) {
	t.Clear()
	fg := t.styles.Table().Header.FgColor.Color()
	bg := t.styles.Table().Header.BgColor.Color()

	var col int
	for _, h := range cdata.Header() {
		if h.Hide || (!t.wide && h.Wide) {
			continue
		}
		if h.Name == "NAMESPACE" && !t.GetModel().ClusterWide() {
			continue
		}
		if h.MX && !t.hasMetrics {
			continue
		}
		if h.VS && vul.ImgScanner == nil {
			continue
		}

		t.AddHeaderCell(col, h)
		c := t.GetCell(0, col)
		c.SetBackgroundColor(bg)
		c.SetTextColor(fg)
		col++
	}
	cdata.Sort(t.getSortCol())

	pads := make(MaxyPad, cdata.HeaderCount())
	ComputeMaxColumns(pads, t.getSortCol().Name, cdata)
	cdata.RowsRange(func(row int, re model1.RowEvent) bool {
		ore, ok := data.FindRow(re.Row.ID)
		if !ok {
			slog.Error("Unable to find original row event", slogs.RowID, re.Row.ID)
			return true
		}
		t.buildRow(row+1, re, ore, cdata.Header(), pads)

		return true
	})

	t.updateSelection(true)
	t.UpdateTitle()
}

func (t *Table) buildRow(r int, re, ore model1.RowEvent, h model1.Header, pads MaxyPad) {
	color := model1.DefaultColorer
	if t.colorerFn != nil {
		color = t.colorerFn
	}

	marked := t.IsMarked(re.Row.ID)
	var col int
	ns := t.GetModel().GetNamespace()
	for c, field := range re.Row.Fields {
		if c >= len(h) {
			slog.Error("Field/header overflow detected. Check your mappings!",
				slogs.GVR, t.GVR(),
				slogs.Cell, c,
				slogs.HeaderSize, len(h),
			)
			continue
		}
		if h[c].Hide || (!t.wide && h[c].Wide) {
			continue
		}

		if h[c].Name == "NAMESPACE" && !t.GetModel().ClusterWide() {
			continue
		}
		if h[c].MX && !t.hasMetrics {
			continue
		}
		if h[c].VS && vul.ImgScanner == nil {
			continue
		}

		if !re.Deltas.IsBlank() && !h.IsTimeCol(c) {
			var old string
			if c < len(ore.Deltas) {
				old = ore.Deltas[c]
			}
			if c < len(re.Deltas) {
				old = re.Deltas[c]
			}
			field += Deltas(old, field)
		}

		if h[c].Decorator != nil {
			field = h[c].Decorator(field)
		}
		if h[c].Align == tview.AlignLeft {
			field = formatCell(field, pads[c])
		}

		cell := tview.NewTableCell(field)
		cell.SetExpansion(1)
		cell.SetAlign(h[c].Align)
		fgColor := color(ns, h, &re)
		cell.SetTextColor(fgColor)
		if marked {
			cell.SetTextColor(t.styles.Table().MarkColor.Color())
		}
		if col == 0 {
			cell.SetReference(re.Row.ID)
		}
		t.SetCell(r, col, cell)
		col++
	}
}

// SortColCmd designates a sorted column.
func (t *Table) SortColCmd(name string, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(*tcell.EventKey) *tcell.EventKey {
		sc := t.getSortCol()
		sc.ASC = !sc.ASC
		if sc.Name != name {
			sc.ASC = asc
		}
		sc.Name = name
		t.setSortCol(sc)
		t.setMSort(true)
		t.Refresh()
		return nil
	}
}

// SortInvertCmd reverses sorting order.
func (t *Table) SortInvertCmd(*tcell.EventKey) *tcell.EventKey {
	t.toggleSortCol()
	t.Refresh()

	return nil
}

// ClearMarks clear out marked items.
func (t *Table) ClearMarks() {
	t.SelectTable.ClearMarks()
	t.Refresh()
}

// Refresh update the table data.
func (t *Table) Refresh() {
	data := t.model.Peek()
	if data.HeaderCount() == 0 {
		return
	}
	// BOZO!! Really want to tell model reload now. Refactor!
	cdata := t.Update(data, t.hasMetrics)
	t.UpdateUI(cdata, data)
}

// GetSelectedRow returns the entire selected row or nil if nothing selected.
func (t *Table) GetSelectedRow(path string) *model1.Row {
	data := t.model.Peek()
	re, ok := data.FindRow(path)
	if !ok {
		return nil
	}

	return &re.Row
}

// NameColIndex returns the index of the resource name column.
func (t *Table) NameColIndex() int {
	col := 0
	if client.IsClusterScoped(t.GetModel().GetNamespace()) {
		return col
	}
	if t.GetModel().ClusterWide() {
		col++
	}

	return col
}

// AddHeaderCell configures a table cell header.
func (t *Table) AddHeaderCell(col int, h model1.HeaderColumn) {
	sc := t.getSortCol()
	sortCol := h.Name == sc.Name
	styles := t.styles.Table()
	c := tview.NewTableCell(sortIndicator(sortCol, sc.ASC, &styles, h.Name))
	c.SetExpansion(1)
	c.SetSelectable(false)
	c.SetAlign(h.Align)
	t.SetCell(0, col, c)
}

func (t *Table) filtered(data *model1.TableData) *model1.TableData {
	return data.Filter(model1.FilterOpts{
		Toast:  t.toast,
		Filter: t.cmdBuff.GetText(),
	})
}

// CmdBuff returns the associated command buffer.
func (t *Table) CmdBuff() *model.FishBuff {
	return t.cmdBuff
}

// ShowDeleted marks row as deleted.
func (t *Table) ShowDeleted() {
	r, _ := t.GetSelection()
	cols := t.GetColumnCount()
	for x := range cols {
		t.GetCell(r, x).SetAttributes(tcell.AttrDim)
	}
}

// UpdateTitle refreshes the table title.
func (t *Table) UpdateTitle() {
	t.SetTitle(t.styleTitle())
}

func (t *Table) styleTitle() string {
	rc := int64(t.GetRowCount())
	if rc > 0 {
		rc--
	}

	ns := t.GetModel().GetNamespace()
	if client.IsClusterWide(ns) || ns == client.NotNamespaced {
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
	if t.Extras != "" {
		ns = t.Extras
	}

	resource := t.gvr.R()
	if t.fullGVR {
		resource = t.gvr.String()
	}

	var (
		title  string
		styles = t.styles.Frame()
	)
	if ns == client.ClusterScope {
		title = SkinTitle(fmt.Sprintf(TitleFmt, resource, render.AsThousands(rc)), &styles)
	} else {
		title = SkinTitle(fmt.Sprintf(NSTitleFmt, resource, ns, render.AsThousands(rc)), &styles)
	}

	buff := t.cmdBuff.GetText()
	if internal.IsLabelSelector(buff) {
		buff = render.Truncate(TrimLabelSelector(buff), maxTruncate)
	} else if l := t.GetModel().GetLabelFilter(); l != "" {
		buff = render.Truncate(l, maxTruncate)
	}

	if buff == "" {
		return title
	}

	return title + SkinTitle(fmt.Sprintf(SearchFmt, buff), &styles)
}

// ROIndicator returns an icon showing whether the session is in readonly mode or not.
func ROIndicator(ro, noIC bool) string {
	switch {
	case noIC:
		return ""
	case ro:
		return lockedIC
	default:
		return unlockedIC
	}
}
