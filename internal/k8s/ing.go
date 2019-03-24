package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Ingress represents a Kubernetes Ingress.
type Ingress struct {
	Connection
}

// NewIngress returns a new Ingress.
func NewIngress(c Connection) Cruder {
	return &Ingress{c}
}

// Get a Ingress.
func (i *Ingress) Get(ns, n string) (interface{}, error) {
	return i.DialOrDie().ExtensionsV1beta1().Ingresses(ns).Get(n, metav1.GetOptions{})
}

// List all Ingresss in a given namespace.
func (i *Ingress) List(ns string) (Collection, error) {
	rr, err := i.DialOrDie().ExtensionsV1beta1().Ingresses(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Ingress.
func (i *Ingress) Delete(ns, n string) error {
	return i.DialOrDie().ExtensionsV1beta1().Ingresses(ns).Delete(n, nil)
}
