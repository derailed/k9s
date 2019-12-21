package views

import "github.com/gdamore/tcell"

type ColumnSortable interface {
	SortColumn(col int, asc bool)
}

func SortColCmd(v ColumnSortable, col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		v.SortColumn(col, asc)
		asc = !asc // flip sort direction for next call

		return nil
	}
}
