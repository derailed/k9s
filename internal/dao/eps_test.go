// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	discoveryv1 "k8s.io/api/discovery/v1"
)

const epsTestNS = "default"

func TestEndpointSlicePodFQNs(t *testing.T) {
	uu := map[string]struct {
		eps *discoveryv1.EndpointSlice
		e   []string
	}{
		"pod_refs": {
			eps: makeEPS(podRef(epsTestNS, "pod-a"), podRef(epsTestNS, "pod-b")),
			e:   []string{"default/pod-a", "default/pod-b"},
		},
		"defaults_namespace": {
			eps: makeEPS(podRef("", "pod-defaulted")),
			e:   []string{"default/pod-defaulted"},
		},
		"cross_namespace": {
			eps: makeEPS(podRef("other", "pod-cross")),
			e:   []string{"other/pod-cross"},
		},
		"skips_non_pod_and_nil_refs": {
			eps: makeEPS(
				podRef(epsTestNS, "pod-kept"),
				&v1.ObjectReference{Kind: "Node", Name: "node-a"},
				nil,
			),
			e: []string{"default/pod-kept"},
		},
		"no_endpoints": {
			eps: makeEPS(),
			e:   []string{},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, endpointSlicePodFQNs(u.eps, epsTestNS))
		})
	}
}

func podRef(ns, name string) *v1.ObjectReference {
	return &v1.ObjectReference{Kind: "Pod", Namespace: ns, Name: name}
}

func makeEPS(refs ...*v1.ObjectReference) *discoveryv1.EndpointSlice {
	eps := discoveryv1.EndpointSlice{
		Endpoints: make([]discoveryv1.Endpoint, 0, len(refs)),
	}
	for _, ref := range refs {
		eps.Endpoints = append(eps.Endpoints, discoveryv1.Endpoint{TargetRef: ref})
	}

	return &eps
}
