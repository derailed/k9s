// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui

import (
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
)

// SelectTable represents a table with selections.
type SelectTable struct {
	*tview.Table

	model      Tabular
	selectedFn func(string) string
	marks      map[string]struct{}
	selFgColor tcell.Color
	selBgColor tcell.Color
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
		if item := s.GetSelectedItem(); item != "" {
			return []string{item}
		}
		return nil
	}

	items := make([]string, 0, len(s.marks))
	for item := range s.marks {
		items = append(items, item)
	}

	return items
}

// GetRowID returns the row id at given location.
func (s *SelectTable) GetRowID(index int) (string, bool) {
	cell := s.GetCell(index, 0)
	if cell == nil {
		return "", false
	}
	id, ok := cell.GetReference().(string)

	return id, ok
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
func (s *SelectTable) SelectRow(r, c int, broadcast bool) {
	if !broadcast {
		s.SetSelectionChangedFunc(nil)
	}
	if c := s.model.RowCount(); c > 0 && r-1 > c {
		r = c + 1
	}
	defer s.SetSelectionChangedFunc(s.selectionChanged)
	s.Select(r, c)
}

// UpdateSelection refresh selected row.
func (s *SelectTable) updateSelection(broadcast bool) {
	r, c := s.GetSelection()
	s.SelectRow(r, c, broadcast)
}

func (s *SelectTable) selectionChanged(r, c int) {
	if r < 0 {
		return
	}
	if cell := s.GetCell(r, c); cell != nil {
		s.SetSelectedStyle(
			tcell.StyleDefault.Foreground(s.selFgColor).
				Background(cell.Color).Attributes(tcell.AttrBold))
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

// ToggleMark toggles marked row.
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

	if cell := s.GetCell(s.GetSelectedRowIndex(), 0); cell != nil {
		s.SetSelectedStyle(tcell.StyleDefault.Foreground(cell.BackgroundColor).Background(cell.Color).Attributes(tcell.AttrBold))
	}
}

// SpanMark toggles marked row.
func (s *SelectTable) SpanMark() {
	selIndex, prev := s.GetSelectedRowIndex(), -1
	if selIndex <= 0 {
		return
	}
	// Look back to find previous mark
	for i := selIndex - 1; i > 0; i-- {
		id, ok := s.GetRowID(i)
		if !ok {
			break
		}
		if _, ok := s.marks[id]; ok {
			prev = i
			break
		}
	}
	if prev != -1 {
		s.markRange(prev, selIndex)
		return
	}

	// Look forward to see if we have a mark
	for i := selIndex; i < s.GetRowCount(); i++ {
		id, ok := s.GetRowID(i)
		if !ok {
			break
		}
		if _, ok := s.marks[id]; ok {
			prev = i
			break
		}
	}
	s.markRange(prev, selIndex)
}

func (s *SelectTable) markRange(prev, curr int) {
	if prev < 0 {
		return
	}
	if prev > curr {
		prev, curr = curr, prev
	}
	for i := prev + 1; i <= curr; i++ {
		id, ok := s.GetRowID(i)
		if !ok {
			break
		}
		s.marks[id] = struct{}{}
		cell := s.GetCell(s.GetSelectedRowIndex(), 0)
		if cell == nil {
			break
		}
		s.SetSelectedStyle(tcell.StyleDefault.Foreground(cell.BackgroundColor).Background(cell.Color).Attributes(tcell.AttrBold))
	}
}

// IsMarked returns true if this item was marked.
func (s *SelectTable) IsMarked(item string) bool {
	_, ok := s.marks[item]
	return ok
}
