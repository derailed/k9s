package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Secret represents a Kubernetes Secret.
type Secret struct{}

// NewSecret returns a new Secret.
func NewSecret() Res {
	return &Secret{}
}

// Get a Secret.
func (c *Secret) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().Secrets(ns).Get(n, opts)
}

// List all Secrets in a given namespace.
func (c *Secret) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().Secrets(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a Secret.
func (c *Secret) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().Secrets(ns).Delete(n, &opts)
}
