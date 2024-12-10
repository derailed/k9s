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
		model1.HeaderColumn{Name: "COMMAND"},
		model1.HeaderColumn{Name: "API-GROUP"},
	}
}

// Render renders a K8s resource to screen.
// BOZO!! Pass in a row with pre-alloc fields??
func (Alias) Render(o interface{}, ns string, r *model1.Row) error {
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
