package render

import (
	"github.com/gdamore/tcell"
)

func rbacVerbHeader() HeaderRow {
	return HeaderRow{
		Header{Name: "GET   "},
		Header{Name: "LIST  "},
		Header{Name: "WATCH "},
		Header{Name: "CREATE"},
		Header{Name: "PATCH "},
		Header{Name: "UPDATE"},
		Header{Name: "DELETE"},
		Header{Name: "DLIST "},
		Header{Name: "EXTRAS"},
	}
}

// Policy renders a rbac policy to screen.
type Policy struct{}

// ColorerFunc colors a resource row.
func (Policy) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Policy) Header(ns string) HeaderRow {
	h := HeaderRow{
		Header{Name: "NAMESPACE"},
		Header{Name: "NAME"},
		Header{Name: "API GROUP"},
		Header{Name: "BINDING"},
	}

	return append(h, rbacVerbHeader()...)
}

// Render renders a K8s resource to screen.
func (Policy) Render(o interface{}, gvr string, r *Row) error {
	panic("NYI")
	return nil
}

// Helpers...
