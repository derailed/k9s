// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultGENHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Generic renders a K8s generic resource to screen.
type Generic struct {
	Base
}

// Header returns a header row.
func (m Generic) Header(_ string) model1.Header {
	return m.doHeader(defaultGENHeader)
}

// Render renders a K8s resource to screen.
func (m Generic) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := m.defaultRow(raw, row); err != nil {
		return err
	}
	if m.specs.isEmpty() {
		return nil
	}
	cols, err := m.specs.realize(o.(*unstructured.Unstructured), defaultGENHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// Render renders a K8s resource to screen.
func (Generic) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	r.ID = client.FQN(raw.GetNamespace(), raw.GetName())
	r.Fields = model1.Fields{
		raw.GetNamespace(),
		raw.GetName(),
		"",
		ToAge(raw.GetCreationTimestamp()),
	}

	return nil
}
