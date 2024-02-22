// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ConfigMap renders a K8s ConfigMap to screen.
type ConfigMap struct {
	Base
}

// Header returns a header rbw.
func (ConfigMap) Header(string) Header {
	return Header{
		HeaderColumn{Name: "NAMESPACE"},
		HeaderColumn{Name: "NAME"},
		HeaderColumn{Name: "DATA"},
		HeaderColumn{Name: "VALID", Wide: true},
		HeaderColumn{Name: "AGE", Time: true},
	}
}

// Render renders a K8s resource to screen.
func (n ConfigMap) Render(o interface{}, _ string, r *Row) error {
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
	r.Fields = Fields{
		cm.Namespace,
		cm.Name,
		strconv.Itoa(len(cm.Data)),
		"",
		ToAge(cm.GetCreationTimestamp()),
	}

	return nil
}
