package render

import (
	"fmt"

	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
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
func (s Subject) Render(o interface{}, ns string, r *Row) error {
	res, ok := o.(SubjectRef)
	if !ok {
		return fmt.Errorf("Expected SubjectRef, but got %T", s)
	}

	r.ID = res.Name
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	r.Fields = append(r.Fields,
		res.Name,
		res.Kind,
		res.FirstLocation,
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

type SubjectRef struct {
	Name, Kind, FirstLocation string
}

// GetObjectKind returns a schema object.
func (SubjectRef) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (s SubjectRef) DeepCopyObject() runtime.Object {
	return s
}
