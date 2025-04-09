// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// Role renders a K8s Role to screen.
type Role struct {
	Base
}

// Header returns a header row.
func (r Role) Header(_ string) model1.Header {
	return r.doHeader(defaultROHeader)
}

var defaultROHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Render renders a K8s resource to screen.
func (r Role) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := r.defaultRow(raw, row); err != nil {
		return err
	}
	if r.specs.isEmpty() {
		return nil
	}
	cols, err := r.specs.realize(raw, defaultROHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (Role) defaultRow(raw *unstructured.Unstructured, row *model1.Row) error {
	var ro rbacv1.Role
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ro)
	if err != nil {
		return err
	}

	row.ID = client.MetaFQN(&ro.ObjectMeta)
	row.Fields = model1.Fields{
		ro.Namespace,
		ro.Name,
		mapToStr(ro.Labels),
		"",
		ToAge(ro.GetCreationTimestamp()),
	}

	return nil
}
