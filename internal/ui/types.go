package ui

import "github.com/derailed/k9s/internal/render"

type (
	// SortFn represent a function that can sort columnar data.
	SortFn func(rows render.Rows, sortCol SortColumn)

	// SortColumn represents a sortable column.
	SortColumn struct {
		index    int
		colCount int
		asc      bool
	}
)
