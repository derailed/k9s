// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultEPsHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "ADDRESSTYPE"},
	model1.HeaderColumn{Name: "PORTS"},
	model1.HeaderColumn{Name: "ENDPOINTS"},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// EndpointSlice renders a K8s EndpointSlice to screen.
type EndpointSlice struct {
	Base
}

// Header returns a header row.
func (e EndpointSlice) Header(_ string) model1.Header {
	return e.doHeader(defaultEPsHeader)
}

// Render renders a K8s resource to screen.
func (e EndpointSlice) Render(o any, ns string, row *model1.Row) error {
	if err := e.defaultRow(o, ns, row); err != nil {
		return err
	}
	if e.specs.isEmpty() {
		return nil
	}
	cols, err := e.specs.realize(o.(*unstructured.Unstructured), defaultEPsHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (e EndpointSlice) defaultRow(o any, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	var eps discoveryv1.EndpointSlice
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &eps)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(&eps.ObjectMeta)
	r.Fields = make(model1.Fields, 0, len(e.Header(ns)))
	r.Fields = model1.Fields{
		eps.Namespace,
		eps.Name,
		string(eps.AddressType),
		toPorts(eps.Ports),
		toEPss(eps.Endpoints),
		ToAge(eps.GetCreationTimestamp()),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toEPss(ee []discoveryv1.Endpoint) string {
	if len(ee) == 0 {
		return UnsetValue
	}

	aa := make([]string, 0, len(ee))
	for _, e := range ee {
		aa = append(aa, e.Addresses...)
	}

	return strings.Join(aa, ",")
}

func toPorts(ee []discoveryv1.EndpointPort) string {
	if len(ee) == 0 {
		return UnsetValue
	}

	aa := make([]string, 0, len(ee))
	for _, e := range ee {
		if e.Port != nil {
			aa = append(aa, strconv.Itoa(int(*e.Port)))
		}
	}

	return strings.Join(aa, ",")
}
