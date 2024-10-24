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

// ConfigMap renders a K8s ConfigMap to screen.
type ConfigMap struct {
	Base
}

// Header returns a header rbw.
func (ConfigMap) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAMESPACE"},
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "DATA"},
		model1.HeaderColumn{Name: "VALID", Wide: true},
		model1.HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (n ConfigMap) Render(o interface{}, _ string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected ConfigMap, but got %T", o)
	}
	var cm v1.ConfigMap
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &cm)
	if err != nil {
		return err
	}

	r.ID = client.FQN(cm.Namespace, cm.Name)
	r.Fields = model1.Fields{
		cm.Namespace,
		cm.Name,
		strconv.Itoa(len(cm.Data)),
		"",
		ToAge(cm.GetCreationTimestamp()),
	}

	return nil
}
