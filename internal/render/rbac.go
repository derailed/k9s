package render

import (
	"github.com/gdamore/tcell"
)

// Rbac renders a rbac to screen.
type Rbac struct{}

// ColorerFunc colors a resource row.
func (Rbac) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Rbac) Header(ns string) HeaderRow {
	h := HeaderRow{
		Header{Name: "NAME"},
		Header{Name: "API GROUP"},
	}

	return append(h, rbacVerbHeader()...)
}

// Render renders a K8s resource to screen.
func (Rbac) Render(o interface{}, gvr string, r *Row) error {
	panic("NYI")
	return nil
}

// Helpers...
