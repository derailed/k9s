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

// ClusterRole renders a K8s ClusterRole to screen.
type ClusterRole struct {
	Base
}

// Header returns a header row.
func (c ClusterRole) Header(_ string) model1.Header {
	return c.doHeader(c.defaultHeader())
}

// Header returns a header rbw.
func (ClusterRole) defaultHeader() model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
		model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
	}
}

// Render renders a K8s resource to screen.
func (p ClusterRole) Render(o interface{}, ns string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expecting clusterrole, but got %T", o)
	}
	if err := p.defaultRow(raw, row); err != nil {
		return err
	}
	if p.specs.isEmpty() {
		return nil
	}

	cols, err := p.specs.realize(raw, p.defaultHeader(), row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

// Render renders a K8s resource to screen.
func (ClusterRole) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	var cr rbacv1.ClusterRole
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cr)
	if err != nil {
		return err
	}

	r.ID = client.FQN("-", cr.ObjectMeta.Name)
	r.Fields = model1.Fields{
		cr.Name,
		mapToStr(cr.Labels),
		ToAge(cr.GetCreationTimestamp()),
	}

	return nil
}
