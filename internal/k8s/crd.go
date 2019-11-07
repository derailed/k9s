package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CustomResourceDefinition represents a Kubernetes CustomResourceDefinition
type CustomResourceDefinition struct {
	*base
	Connection
}

// NewCustomResourceDefinition returns a new CustomResourceDefinition.
func NewCustomResourceDefinition(c Connection) *CustomResourceDefinition {
	return &CustomResourceDefinition{&base{}, c}
}

// Get a CustomResourceDefinition.
func (c *CustomResourceDefinition) Get(_, n string) (interface{}, error) {
	return c.NSDialOrDie().Get(n, metav1.GetOptions{})
}

// List all CustomResourceDefinitions in a given namespace.
func (c *CustomResourceDefinition) List(_ string, opts metav1.ListOptions) (Collection, error) {
	rr, err := c.NSDialOrDie().List(opts)
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a CustomResourceDefinition.
func (c *CustomResourceDefinition) Delete(_, n string, cascade, force bool) error {
	return c.NSDialOrDie().Delete(n, nil)
}
