// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Secret renders a K8s Secret to screen.
type Secret struct {
	Base
}

// Header returns a header row.
func (s Secret) Header(_ string) model1.Header {
	return s.doHeader(s.defaultHeader())
}

func (Secret) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "TYPE"},
		model1.HeaderColumn{Name: "DATA"},
		model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (s Secret) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Secret, but got %T", o)
	}

	if err := s.defaultRow(raw, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}

	cols, err := s.specs.realize(raw, s.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (n Secret) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var sec v1.Secret
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sec)
	if err != nil {
		return err
	}

	r.ID = client.FQN(sec.Namespace, sec.Name)
	r.Fields = model1.Fields{
		sec.Namespace,
		sec.Name,
		string(sec.Type),
		strconv.Itoa(len(sec.Data)),
		"",
		ToAge(raw.GetCreationTimestamp()),
	}

	return nil
}
