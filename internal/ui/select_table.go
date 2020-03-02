package ui

import (
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// SelectTable represents a table with selections.
type SelectTable struct {
	*tview.Table

	model      Tabular
	selectedFn func(string) string
	marks      map[string]struct{}
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
	for item := range s.marks {
		items = append(items, item)
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
	r, _ := s.GetSelection()
	return TrimCell(s, r, col)
}

// SetSelectedFn defines a function that cleanse the current selection.
func (s *SelectTable) SetSelectedFn(f func(string) string) {
	s.selectedFn = f
}

// GetSelectedRowIndex fetch the currently selected row index.
func (s *SelectTable) GetSelectedRowIndex() int {
	r, _ := s.GetSelection()
	return r
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
	r, _ := s.GetSelection()
	s.SelectRow(r, broadcast)
}

func (s *SelectTable) selectionChanged(r, c int) {
	if r < 0 {
		return
	}
	cell := s.GetCell(r, c)
	s.SetSelectedStyle(tcell.ColorBlack, cell.Color, tcell.AttrBold)
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
	sel := s.GetSelectedItem()
	if sel == "" {
		return
	}
	if _, ok := s.marks[sel]; ok {
		delete(s.marks, s.GetSelectedItem())
	} else {
		s.marks[sel] = struct{}{}
	}

	cell := s.GetCell(s.GetSelectedRowIndex(), 0)
	s.SetSelectedStyle(
		tcell.ColorBlack,
		cell.Color,
		tcell.AttrBold,
	)
}

// IsMarked returns true if this item was marked.
func (s *Table) IsMarked(item string) bool {
	_, ok := s.marks[item]
	return ok
}
