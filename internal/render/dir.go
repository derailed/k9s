// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Dir renders a directory entry to screen.
type Dir struct{}

// IsGeneric identifies a generic handler.
func (Dir) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (Dir) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorCadetBlue
	}
}

// Header returns a header row.
func (Dir) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (Dir) Render(o interface{}, ns string, r *model1.Row) error {
	d, ok := o.(DirRes)
	if !ok {
		return fmt.Errorf("expected DirRes, but got %T", o)
	}

	name := "🦄 "
	if d.Entry.IsDir() {
		name = "📁 "
	}
	name += d.Entry.Name()
	r.ID, r.Fields = d.Path, append(r.Fields, name)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// DirRes represents an alias resource.
type DirRes struct {
	Entry os.DirEntry
	Path  string
}

// GetObjectKind returns a schema object.
func (DirRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (d DirRes) DeepCopyObject() runtime.Object {
	return d
}
