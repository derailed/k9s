// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var defaultEPHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "ENDPOINTS"},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// Endpoints renders a K8s Endpoints to screen.
type Endpoints struct {
	Base
}

// Header returns a header row.
func (e Endpoints) Header(_ string) model1.Header {
	return e.doHeader(defaultEPHeader)
}

// Render renders a K8s resource to screen.
func (e Endpoints) Render(o any, ns string, row *model1.Row) error {
	if err := e.defaultRow(o, ns, row); err != nil {
		return err
	}
	if e.specs.isEmpty() {
		return nil
	}
	cols, err := e.specs.realize(o.(*unstructured.Unstructured), defaultEPHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (e Endpoints) defaultRow(o any, ns string, r *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	var ep v1.Endpoints
	err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw.Object, &ep)
	if err != nil {
		return err
	}

	r.ID = client.MetaFQN(&ep.ObjectMeta)
	r.Fields = make(model1.Fields, 0, len(e.Header(ns)))
	r.Fields = model1.Fields{
		ep.Namespace,
		ep.Name,
		missing(toEPs(ep.Subsets)),
		ToAge(ep.GetCreationTimestamp()),
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

func toEPs(ss []v1.EndpointSubset) string {
	aa := make([]string, 0, len(ss))
	for _, s := range ss {
		pp := make([]string, len(s.Ports))
		portsToStrs(s.Ports, pp)
		a := make([]string, len(s.Addresses))
		processIPs(a, pp, s.Addresses)
		aa = append(aa, strings.Join(a, ","))
	}
	return strings.Join(aa, ",")
}

func portsToStrs(pp []v1.EndpointPort, ss []string) {
	for i := range pp {
		ss[i] = strconv.Itoa(int(pp[i].Port))
	}
}

func processIPs(aa, pp []string, addrs []v1.EndpointAddress) {
	const maxIPs = 3
	var i int
	for _, a := range addrs {
		if a.IP == "" {
			continue
		}
		if len(pp) == 0 {
			aa[i], i = a.IP, i+1
			continue
		}
		if len(pp) > maxIPs {
			aa[i], i = a.IP+":"+strings.Join(pp[:maxIPs], ",")+"...", i+1
		} else {
			aa[i], i = a.IP+":"+strings.Join(pp, ","), i+1
		}
	}
}
