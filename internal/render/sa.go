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

// ServiceAccount renders a K8s ServiceAccount to screen.
type ServiceAccount struct {
	Base
}

// Header returns a header row.
func (ServiceAccount) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "SECRET"},
		model1.HeaderColumn{Name: "LABELS", Wide: true},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (s ServiceAccount) Render(o interface{}, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected ServiceAccount, but got %T", o)
	}
	var sa v1.ServiceAccount
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &sa)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(sa.ObjectMeta)
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
