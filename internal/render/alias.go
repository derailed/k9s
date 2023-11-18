// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Alias renders a aliases to screen.
type Alias struct {
	Base
}

// Header returns a header row.
func (Alias) Header(ns string) Header {
	return Header{
		HeaderColumn{Name: "RESOURCE"},
		HeaderColumn{Name: "COMMAND"},
		HeaderColumn{Name: "APIGROUP"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (Alias) Render(o interface{}, ns string, r *Row) error {
	a, ok := o.(AliasRes)
	if !ok {
		return fmt.Errorf("expected AliasRes, but got %T", o)
	}

	r.ID = a.GVR
	gvr := client.NewGVR(a.GVR)
	res, grp := gvr.RG()
	r.Fields = append(r.Fields,
		res,
		strings.Join(a.Aliases, ","),
		grp,
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
