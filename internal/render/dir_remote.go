// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// DirRemote renders a directory entry to screen.
type DirRemote struct{}

// IsGeneric identifies a generic handler.
func (DirRemote) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (DirRemote) ColorerFunc() model1.ColorerFunc {
	return func(ns string, _ model1.Header, re *model1.RowEvent) tcell.Color {
		return tcell.ColorCadetBlue
	}
}

// Header returns a header row.
func (DirRemote) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (DirRemote) Render(o interface{}, ns string, r *model1.Row) error {
	d, ok := o.(DirRemoteRes)
	if !ok {
		return fmt.Errorf("expected DirRemoteRes, but got %T", o)
	}
	var name string
	var path = d.Path
	if strings.HasSuffix(d.Name, "/") { // directory
		name = "📁 " + strings.TrimSuffix(d.Name, "/")
		path = path + "/" // was stripped by filepath
	} else {
		name = "📄 " + d.Name
	}
	r.ID, r.Fields = path, append(r.Fields, name)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// DirRes represents an alias resource.
type DirRemoteRes struct {
	Name string
	Path string
}

// GetObjectKind returns a schema object.
func (DirRemoteRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (d DirRemoteRes) DeepCopyObject() runtime.Object {
	return d
}
