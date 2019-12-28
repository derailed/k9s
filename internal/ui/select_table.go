package ui

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// Namespaceable represents a namespaceable model.
type Namespaceable interface {
	// ClusterWide returns true if the model represents resource in all namespaces.
	ClusterWide() bool

	// GetNamespace returns the model namespace.
	GetNamespace() string

	// SetNamespace changes the model namespace.
	SetNamespace(string)

	// InNamespace check if current namespace matches models.
	InNamespace(string) bool
}

// Tabular represents a tabular model.
type Tabular interface {
	Namespaceable

	// Empty returns true if model has no data.
	Empty() bool

	// Peek returns current model data.
	Peek() render.TableData

	// Watch watches a given resource for changes.
	Watch(context.Context)

	// SetRefreshRate sets the model watch loop rate.
	SetRefreshRate(time.Duration)

	// AddListener registers a model listener.
	AddListener(model.TableListener)
}

// Selectable represents a table with selections.
type SelectTable struct {
	*tview.Table

	model              Tabular
	selectedRow        int
	selectedFn         func(string) string
	selectionListeners []SelectedRowFunc
	marks              map[string]bool
}

// SetModel sets the table model.
func (s *SelectTable) SetModel(m Tabular) {
	s.model = m
}

// GetModel returns the current model.
func (s *SelectTable) GetModel() Tabular {
	return s.model
}

// ClearSelection reset selected row.
func (s *SelectTable) ClearSelection() {
	s.Select(0, 0)
	s.ScrollToBeginning()
}

// SelectFirstRow select first data row if any.
func (s *SelectTable) SelectFirstRow() {
	if s.GetRowCount() > 0 {
		s.Select(1, 0)
	}
}

// GetSelectedItems return currently marked or selected items names.
func (s *SelectTable) GetSelectedItems() []string {
	if len(s.marks) == 0 {
		return []string{s.GetSelectedItem()}
	}

	var items []string
	for item, marked := range s.marks {
		if marked {
			items = append(items, item)
		}
	}

	return items
}

// GetSelectedItem returns the currently selected item name.
func (s *SelectTable) GetSelectedItem() string {
	if s.GetSelectedRowIndex() == 0 || s.model.Empty() {
		return ""
	}
	sel, ok := s.GetCell(s.GetSelectedRowIndex(), 0).GetReference().(string)
	if !ok {
		return ""
	}
	if s.selectedFn != nil {
		return s.selectedFn(sel)
	}
	return sel
}

// GetSelectedCell returns the content of a cell for the currently selected row.
func (s *SelectTable) GetSelectedCell(col int) string {
	return TrimCell(s, s.selectedRow, col)
}

// SetSelectedFn defines a function that cleanse the current selection.
func (s *SelectTable) SetSelectedFn(f func(string) string) {
	s.selectedFn = f
}

// GetSelectedRow fetch the currently selected row index.
func (s *SelectTable) GetSelectedRowIndex() int {
	return s.selectedRow
}

// SelectRow select a given row by index.
func (s *SelectTable) SelectRow(r int, broadcast bool) {
	if !broadcast {
		s.SetSelectionChangedFunc(nil)
	}
	defer s.SetSelectionChangedFunc(s.selectionChanged)
	s.Select(r, 0)
}

// UpdateSelection refresh selected row.
func (s *SelectTable) updateSelection(broadcast bool) {
	s.SelectRow(s.selectedRow, broadcast)
}

func (s *SelectTable) selectionChanged(r, c int) {
	s.selectedRow = r
	if r == 0 {
		return
	}

	if s.marks[s.GetSelectedItem()] {
		s.SetSelectedStyle(tcell.ColorBlack, tcell.ColorCadetBlue, tcell.AttrBold)
	} else {
		cell := s.GetCell(r, c)
		s.SetSelectedStyle(tcell.ColorBlack, cell.Color, tcell.AttrBold)
	}

	for _, f := range s.selectionListeners {
		f(r)
	}
}

// ClearMarks delete all marked items.
func (s *SelectTable) ClearMarks() {
	for k := range s.marks {
		delete(s.marks, k)
	}
}

// DeleteMark delete a marked item.
func (s *SelectTable) DeleteMark(k string) {
	delete(s.marks, k)
}

// ToggleMark toggles marked row
func (s *SelectTable) ToggleMark() {
	s.marks[s.GetSelectedItem()] = !s.marks[s.GetSelectedItem()]
	if !s.marks[s.GetSelectedItem()] {
		return
	}

	cell := s.GetCell(s.GetSelectedRowIndex(), 0)
	s.SetSelectedStyle(
		tcell.ColorBlack,
		cell.Color,
		tcell.AttrBold,
	)
}

func (s *Table) IsMarked(item string) bool {
	return s.marks[item]
}

// AddSelectedRowListener add a new selected row listener.
func (s *SelectTable) AddSelectedRowListener(f SelectedRowFunc) {
	s.selectionListeners = append(s.selectionListeners, f)
}
