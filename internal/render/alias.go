// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Alias renders an aliases to screen.
type Alias struct {
	Base
}

// Header returns a header row.
func (Alias) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "RESOURCE"},
		model1.HeaderColumn{Name: "GROUP"},
		model1.HeaderColumn{Name: "VERSION"},
		model1.HeaderColumn{Name: "COMMAND"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (Alias) Render(o interface{}, ns string, r *model1.Row) error {
	a, ok := o.(AliasRes)
	if !ok {
		return fmt.Errorf("expected AliasRes, but got %T", o)
	}

	gvr := client.NewGVR(a.GVR)
	r.ID = gvr.String()
	r.Fields = append(r.Fields,
		gvr.R(),
		gvr.G(),
		gvr.V(),
		strings.Join(a.Aliases, ","),
	)

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// AliasRes represents an alias resource.
type AliasRes struct {
	GVR     string
	Aliases []string
}

// GetObjectKind returns a schema object.
func (AliasRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (a AliasRes) DeepCopyObject() runtime.Object {
	return a
}
