package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DaemonSet represents a Kubernetes DaemonSet
type DaemonSet struct{}

// NewDaemonSet returns a new DaemonSet.
func NewDaemonSet() Res {
	return &DaemonSet{}
}

// Get a DaemonSet.
func (*DaemonSet) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().ExtensionsV1beta1().DaemonSets(ns).Get(n, opts)
}

// List all DaemonSets in a given namespace.
func (*DaemonSet) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().ExtensionsV1beta1().DaemonSets(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a DaemonSet.
func (*DaemonSet) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().ExtensionsV1beta1().DaemonSets(ns).Delete(n, &opts)
}
