package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ingress represents a Kubernetes Ingress.
type Ingress struct{}

// NewIngress returns a new Ingress.
func NewIngress() Res {
	return &Ingress{}
}

// Get a Ingress.
func (*Ingress) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().ExtensionsV1beta1().Ingresses(ns).Get(n, opts)
}

// List all Ingresss in a given namespace.
func (*Ingress) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().ExtensionsV1beta1().Ingresses(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Ingress.
func (*Ingress) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().ExtensionsV1beta1().Ingresses(ns).Delete(n, &opts)
}
