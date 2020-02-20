package render

import (
	"fmt"

	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Subject renders a rbac to screen.
type Subject struct{}

// Happy returns true if resoure is happy, false otherwise
func (Subject) Happy(_ string, _ Row) bool {
	return true
}

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
		Header{Name: "VALID", Wide: true},
	}
}

// Render renders a K8s resource to screen.
func (s Subject) Render(o interface{}, ns string, r *Row) error {
	res, ok := o.(SubjectRes)
	if !ok {
		return fmt.Errorf("Expected SubjectRes, but got %T", s)
	}

	r.ID = res.Name
	r.Fields = make(Fields, 0, len(s.Header(ns)))
	r.Fields = append(r.Fields,
		res.Name,
		res.Kind,
		res.FirstLocation,
		"",
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// SubjectRes represents a subject rule.
type SubjectRes struct {
	Name, Kind, FirstLocation string
}

// GetObjectKind returns a schema object.
func (SubjectRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (s SubjectRes) DeepCopyObject() runtime.Object {
	return s
}

// Subjects represents a collection of RBAC policies.
type Subjects []SubjectRes

// Upsert adds a new subject.
func (ss Subjects) Upsert(s SubjectRes) Subjects {
	idx, ok := ss.find(s.Name)
	if !ok {
		return append(ss, s)
	}
	ss[idx] = s

	return ss
}

// Find locates a row by id. Retturns false is not found.
func (ss Subjects) find(res string) (int, bool) {
	for i, s := range ss {
		if s.Name == res {
			return i, true
		}
	}

	return 0, false
}
