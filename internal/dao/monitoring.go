// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor = (*ServiceMonitor)(nil)
	_ Accessor = (*PodMonitor)(nil)
	_ Accessor = (*Probe)(nil)
)

// MonitoringTarget describes the set of resources selected by a Prometheus
// Operator monitoring CRD. An empty Namespace means "search all namespaces";
// the label selector is then relied on to narrow the results.
type MonitoringTarget struct {
	Namespace string
	Selector  labels.Selector
}

// ServiceMonitor represents a Prometheus Operator ServiceMonitor CRD.
type ServiceMonitor struct {
	Resource
}

// PodMonitor represents a Prometheus Operator PodMonitor CRD.
type PodMonitor struct {
	Resource
}

// Probe represents a Prometheus Operator Probe CRD.
type Probe struct {
	Resource
}

// SelectedTarget returns the Services selected by the given ServiceMonitor.
func (sm *ServiceMonitor) SelectedTarget(fqn string) (*MonitoringTarget, error) {
	u, err := getUnstructured(sm.getFactory(), sm.gvr, fqn)
	if err != nil {
		return nil, err
	}
	return targetFromSpec(u, "spec")
}

// SelectedTarget returns the Pods selected by the given PodMonitor.
func (pm *PodMonitor) SelectedTarget(fqn string) (*MonitoringTarget, error) {
	u, err := getUnstructured(pm.getFactory(), pm.gvr, fqn)
	if err != nil {
		return nil, err
	}
	return targetFromSpec(u, "spec")
}

// SelectedTarget returns the Ingresses selected by the given Probe. Returns
// (nil, nil) when the Probe has no ingress selector (i.e. staticConfig-only).
func (p *Probe) SelectedTarget(fqn string) (*MonitoringTarget, error) {
	u, err := getUnstructured(p.getFactory(), p.gvr, fqn)
	if err != nil {
		return nil, err
	}
	if _, found, _ := unstructured.NestedMap(u.Object, "spec", "targets", "ingress"); !found {
		return nil, nil
	}
	return targetFromSpec(u, "spec", "targets", "ingress")
}

func getUnstructured(f Factory, gvr *client.GVR, fqn string) (*unstructured.Unstructured, error) {
	o, err := f.Get(gvr, fqn, true, labels.Everything())
	if err != nil {
		return nil, err
	}
	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("expected unstructured object for %s", fqn)
	}
	return u, nil
}

// targetFromSpec builds a MonitoringTarget from a nested location on the
// unstructured object. specPath ends at the parent map that carries both a
// "selector" (metav1.LabelSelector) and a "namespaceSelector" (Prometheus
// Operator's NamespaceSelector: {any, matchNames}).
func targetFromSpec(u *unstructured.Unstructured, specPath ...string) (*MonitoringTarget, error) {
	labelSel, err := labelSelectorAt(u, append(specPath, "selector")...)
	if err != nil {
		return nil, err
	}
	sel, err := metav1.LabelSelectorAsSelector(labelSel)
	if err != nil {
		return nil, fmt.Errorf("invalid selector: %w", err)
	}

	nsPath := append(specPath, "namespaceSelector")
	anyNS, _, err := unstructured.NestedBool(u.Object, append(nsPath, "any")...)
	if err != nil {
		return nil, fmt.Errorf("invalid namespaceSelector.any: %w", err)
	}
	names, _, err := unstructured.NestedStringSlice(u.Object, append(nsPath, "matchNames")...)
	if err != nil {
		return nil, fmt.Errorf("invalid namespaceSelector.matchNames: %w", err)
	}

	return &MonitoringTarget{
		Namespace: namespaceFromSelector(u.GetNamespace(), anyNS, names),
		Selector:  sel,
	}, nil
}

// labelSelectorAt reads a metav1.LabelSelector out of the given unstructured
// path. A missing selector is treated as the empty selector (matches nothing);
// this mirrors how metav1.LabelSelectorAsSelector treats a nil pointer.
func labelSelectorAt(u *unstructured.Unstructured, path ...string) (*metav1.LabelSelector, error) {
	raw, found, err := unstructured.NestedMap(u.Object, path...)
	if err != nil {
		return nil, fmt.Errorf("invalid selector at %v: %w", path, err)
	}
	if !found {
		return &metav1.LabelSelector{}, nil
	}
	var sel metav1.LabelSelector
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(raw, &sel); err != nil {
		return nil, errors.New("selector is not a valid LabelSelector")
	}
	return &sel, nil
}

// namespaceFromSelector resolves a Prometheus Operator NamespaceSelector to a
// single k9s active-namespace value. Because k9s' active-namespace model is
// single-namespace-or-all, a matchNames list with more than one entry is
// widened to all-namespaces; the label selector still filters the results.
func namespaceFromSelector(ownNS string, anyNS bool, matchNames []string) string {
	switch {
	case anyNS:
		return client.NamespaceAll
	case len(matchNames) == 1:
		return matchNames[0]
	case len(matchNames) > 1:
		return client.NamespaceAll
	default:
		return ownNS
	}
}
