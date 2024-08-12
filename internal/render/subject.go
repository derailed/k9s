// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Subject renders a rbac to screen.
type Subject struct {
	Base
}

// ColorerFunc colors a resource row.
func (Subject) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorMediumSpringGreen
	}
}

// Header returns a header row.
func (Subject) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "KIND"},
		model1.HeaderColumn{Name: "FIRST LOCATION"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
	}
}

// Render renders a K8s resource to screen.
func (s Subject) Render(o interface{}, ns string, r *model1.Row) error {
	res, ok := o.(SubjectRes)
	if !ok {
		return fmt.Errorf("expected SubjectRes, but got %T", s)
	}

	r.ID = res.Name
	r.Fields = model1.Fields{
		res.Name,
		res.Kind,
		res.FirstLocation,
		"",
	}

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

// Find locates a row by id. Returns false is not found.
func (ss Subjects) find(res string) (int, bool) {
	for i, s := range ss {
		if s.Name == res {
			return i, true
		}
	}

	return 0, false
}
