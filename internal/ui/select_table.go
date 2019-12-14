package ui

import (
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
)

// Selectable represents a table with selections.
type SelectTable struct {
	*tview.Table

	Data         render.TableData
	selectedItem string
	selectedRow  int
	selectedFn   func(string) string
	selListeners []SelectedRowFunc
	marks        map[string]bool
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
	if s.selectedFn != nil {
		return s.selectedFn(s.selectedItem)
	}
	return s.selectedItem
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

// RowSelected checks if there is an active row selection.
func (s *SelectTable) RowSelected() bool {
	return s.selectedItem != ""
}

// GetRow retrieves the entire selected row.
func (s *SelectTable) GetRow() render.Row {
	return s.Data.RowEvents[s.GetSelectedRowIndex()].Row
}

func (s *SelectTable) updateSelectedItem(r int) {
	if r <= 0 || len(s.Data.RowEvents) == 0 {
		s.selectedItem = ""
		return
	}

	if r-1 >= len(s.Data.RowEvents) {
		return
	}
	s.selectedItem = s.Data.RowEvents[r-1].Row.ID
}

// SelectRow select a given row by index.
func (s *SelectTable) SelectRow(r int, broadcast bool) {
	if !broadcast {
		s.SetSelectionChangedFunc(nil)
	}
	defer s.SetSelectionChangedFunc(s.selChanged)
	s.Select(r, 0)
	s.updateSelectedItem(r)
}

// UpdateSelection refresh selected row.
func (s *SelectTable) updateSelection(broadcast bool) {
	s.SelectRow(s.selectedRow, broadcast)
}

func (s *SelectTable) selChanged(r, c int) {
	s.selectedRow = r
	s.updateSelectedItem(r)
	if r == 0 {
		return
	}

	if s.marks[s.GetSelectedItem()] {
		s.SetSelectedStyle(tcell.ColorBlack, tcell.ColorCadetBlue, tcell.AttrBold)
	} else {
		cell := s.GetCell(r, c)
		s.SetSelectedStyle(tcell.ColorBlack, cell.Color, tcell.AttrBold)
	}

	for _, f := range s.selListeners {
		f(r, c)
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
	s.selListeners = append(s.selListeners, f)
}
