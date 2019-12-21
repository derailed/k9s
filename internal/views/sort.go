package views

import "github.com/gdamore/tcell"

type columnSortable interface {
	sortColumn(col int, asc bool)
}

func sortColCmd(v columnSortable, col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		v.sortColumn(col, asc)
		asc = !asc // flip sort direction for next call

		return nil
	}
}
