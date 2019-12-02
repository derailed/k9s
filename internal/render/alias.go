package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/gdamore/tcell"
)

// Alias renders a aliases to screen.
type Alias struct{}

// ColorerFunc colors a resource row.
func (Alias) ColorerFunc() ColorerFunc {
	return func(ns string, re RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Alias) Header(ns string) HeaderRow {
	return HeaderRow{
		Header{Name: "RESOURCE"},
		Header{Name: "COMMAND"},
		Header{Name: "APIGROUP"},
	}
}

// Render renders a K8s resource to screen.
func (Alias) Render(o interface{}, gvr string, r *Row) error {
	aliases, ok := o.([]string)
	if !ok {
		return fmt.Errorf("Expected Alias, but got %T", o)
	}

	g := k8s.GVR(gvr)
	r.ID = string(gvr)
	r.Fields = Fields{
		g.ToR(),
		strings.Join(aliases, ","),
		g.ToG(),
		// Pad(g.ToR(), 30),
		// Pad(strings.Join(aliases, ","), 70),
		// Pad(g.ToG(), 30),
	}

	return nil
}

// Helpers...

// Pad a string up to the given length or truncates if greater than length.
func Pad(s string, width int) string {
	if len(s) == width {
		return s
	}

	if len(s) > width {
		return Truncate(s, width)
	}

	return s + strings.Repeat(" ", width-len(s))
}
