package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret represents a Kubernetes Secret.
type Secret struct {
	Connection
}

// NewSecret returns a new Secret.
func NewSecret(c Connection) Cruder {
	return &Secret{c}
}

// Get a Secret.
func (s *Secret) Get(ns, n string) (interface{}, error) {
	return s.DialOrDie().CoreV1().Secrets(ns).Get(n, metav1.GetOptions{})
}

// List all Secrets in a given namespace.
func (s *Secret) List(ns string) (Collection, error) {
	rr, err := s.DialOrDie().CoreV1().Secrets(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Secret.
func (s *Secret) Delete(ns, n string) error {
	return s.DialOrDie().CoreV1().Secrets(ns).Delete(n, nil)
}
