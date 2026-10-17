// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"log/slog"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	discoveryv1 "k8s.io/api/discovery/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
)

// podsForService resolves the pods backing a service via its EndpointSlices
func podsForService(f Factory, ns, svcName string) (sets.Set[string], error) {
	pods, err := podsFromEndpointSlices(f, ns, svcName)
	if err != nil {
		slog.Warn("Endpointslice lookup failed; falling back to service selector",
			slogs.Namespace, ns,
			slogs.ResName, svcName,
			slogs.Error, err,
		)
	}
	if pods.Len() > 0 {
		return pods, nil
	}

	return podsFromServiceSelector(f, ns, svcName)
}

// podsFromEndpointSlices collects the FQNs of pods referenced by a service's
// EndpointSlices through their endpoint targetRefs.
func podsFromEndpointSlices(f Factory, ns, svcName string) (sets.Set[string], error) {
	sel := labels.Set{discoveryv1.LabelServiceName: svcName}.AsSelector()
	oo, err := f.List(client.EpsGVR, ns, true, sel)
	if err != nil {
		return nil, err
	}

	pods := sets.New[string]()
	for _, o := range oo {
		var eps discoveryv1.EndpointSlice
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		if err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &eps); err != nil {
			return nil, err
		}
		pods.Insert(endpointSlicePodFQNs(&eps, ns)...)
	}

	return pods, nil
}

// endpointSlicePodFQNs returns the FQNs of pods referenced by an EndpointSlice,
// defaulting the namespace to the slice's own when a targetRef omits it.
func endpointSlicePodFQNs(eps *discoveryv1.EndpointSlice, defaultNS string) []string {
	fqns := make([]string, 0, len(eps.Endpoints))
	for i := range eps.Endpoints {
		ref := eps.Endpoints[i].TargetRef
		if ref == nil || ref.Kind != "Pod" {
			continue
		}
		refNS := ref.Namespace
		if refNS == "" {
			refNS = defaultNS
		}
		fqns = append(fqns, client.FQN(refNS, ref.Name))
	}

	return fqns
}

// podsFromServiceSelector collects the FQNs of pods matching a service's
// selector, if no pods are resolved by EndpointSlices (eg, no RBAC)
func podsFromServiceSelector(f Factory, ns, svcName string) (sets.Set[string], error) {
	var svc Service
	svc.Init(f, client.SvcGVR)
	instance, err := svc.GetInstance(client.FQN(ns, svcName))
	if err != nil {
		return nil, err
	}

	pods := sets.New[string]()
	if len(instance.Spec.Selector) == 0 {
		return pods, nil
	}

	oo, err := f.List(client.PodGVR, ns, true, labels.Set(instance.Spec.Selector).AsSelector())
	if err != nil {
		return nil, err
	}
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			continue
		}
		pods.Insert(client.FQN(u.GetNamespace(), u.GetName()))
	}

	return pods, nil
}
