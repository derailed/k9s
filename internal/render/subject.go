package render

import (
	"github.com/gdamore/tcell"
)

// Subject renders a rbac to screen.
type Subject struct{}

// ColorerFunc colors a resource row.
func (Subject) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Subject) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "KIND"},
		Header{Name: "FIRST LOCATION"},
	}
}

// Render renders a K8s resource to screen.
func (Subject) Render(o interface{}, gvr string, r *Row) error {
	return nil
}
