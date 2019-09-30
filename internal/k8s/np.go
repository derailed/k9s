package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicy represents a Kubernetes NetworkPolicy
type NetworkPolicy struct {
	*Resource
	Connection
}

// NewNetworkPolicy returns a new NetworkPolicy.
func NewNetworkPolicy(c Connection, gvr GVR) *NetworkPolicy {
	return &NetworkPolicy{&Resource{gvr: gvr}, c}
}

// Get a NetworkPolicy.
func (d *NetworkPolicy) Get(ns, n string) (interface{}, error) {
	return d.DialOrDie().ExtensionsV1beta1().NetworkPolicies(ns).Get(n, metav1.GetOptions{})
}

// List all NetworkPolicys in a given namespace.
func (d *NetworkPolicy) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: d.labelSelector,
		FieldSelector: d.fieldSelector,
	}
	rr, err := d.DialOrDie().ExtensionsV1beta1().NetworkPolicies(ns).List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a NetworkPolicy.
func (d *NetworkPolicy) Delete(ns, n string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	return d.DialOrDie().ExtensionsV1beta1().NetworkPolicies(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}
