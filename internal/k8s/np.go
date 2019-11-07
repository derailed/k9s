package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetworkPolicy represents a Kubernetes NetworkPolicy
type NetworkPolicy struct {
	*base
	Connection
}

// NewNetworkPolicy returns a new NetworkPolicy.
func NewNetworkPolicy(c Connection) *NetworkPolicy {
	return &NetworkPolicy{&base{}, c}
}

// Get a NetworkPolicy.
func (d *NetworkPolicy) Get(ns, n string) (interface{}, error) {
	return d.DialOrDie().NetworkingV1().NetworkPolicies(ns).Get(n, metav1.GetOptions{})
}

// List all NetworkPolicys in a given namespace.
func (d *NetworkPolicy) List(ns string, opts metav1.ListOptions) (Collection, error) {
	rr, err := d.DialOrDie().NetworkingV1().NetworkPolicies(ns).List(opts)
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
	return d.DialOrDie().NetworkingV1().NetworkPolicies(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}
