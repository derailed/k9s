// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultSECHeader = model1.Header{
	model1.HeaderColumn{Name: colNamespace},
	model1.HeaderColumn{Name: colName},
	model1.HeaderColumn{Name: colAge, Attrs: model1.Attrs{Time: true}},
}

// Secret renders a K8s Secret to screen.
type Secret struct {
	Base
}

// Header returns a header row.
func (s Secret) Header(_ string) model1.Header {
	return s.doHeader(defaultSECHeader)
}

// Render renders a K8s resource to screen.
func (s Secret) Render(o any, _ string, row *model1.Row) error {
	obj, ok := o.(runtime.Object)
	if !ok {
		return fmt.Errorf("expected Secret metadata, but got %T", o)
	}
	if err := s.defaultRow(o, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}
	cols, err := s.specs.realize(obj, defaultSECHeader, row)
	cols.hydrateRow(row)

	return err
}

func (Secret) defaultRow(o any, r *model1.Row) error {
	var meta metav1.Object
	switch obj := o.(type) {
	case *metav1.PartialObjectMetadata:
		meta = obj
	case *unstructured.Unstructured:
		meta = obj
	default:
		return fmt.Errorf("expected Secret metadata, but got %T", o)
	}

	r.ID = client.FQN(meta.GetNamespace(), meta.GetName())
	r.Fields = model1.Fields{
		meta.GetNamespace(),
		meta.GetName(),
		ToAge(meta.GetCreationTimestamp()),
	}

	return nil
}
