// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultSAHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "SECRET"},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// ServiceAccount renders a K8s ServiceAccount to screen.
type ServiceAccount struct {
	Base
}

// Header returns a header row.
func (s ServiceAccount) Header(_ string) model1.Header {
	return s.doHeader(defaultSAHeader)
}

// Render renders a K8s resource to screen.
func (s ServiceAccount) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := s.defaultRow(raw, row); err != nil {
		return err
	}
	if s.specs.isEmpty() {
		return nil
	}
	cols, err := s.specs.realize(raw, defaultSAHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (ServiceAccount) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var sa v1.ServiceAccount
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(&sa.ObjectMeta)
	r.Fields = model1.Fields{
		sa.Namespace,
		sa.Name,
		strconv.Itoa(len(sa.Secrets)),
		mapToStr(sa.Labels),
		"",
		ToAge(sa.GetCreationTimestamp()),
	}

	return nil
}
