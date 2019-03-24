package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMap represents a Kubernetes ConfigMap
type ConfigMap struct {
	Connection
}

// NewConfigMap returns a new ConfigMap.
func NewConfigMap(c Connection) Cruder {
	return &ConfigMap{c}
}

// Get a ConfigMap.
func (c *ConfigMap) Get(ns, n string) (interface{}, error) {
	return c.DialOrDie().CoreV1().ConfigMaps(ns).Get(n, metav1.GetOptions{})
}

// List all ConfigMaps in a given namespace.
func (c *ConfigMap) List(ns string) (Collection, error) {
	rr, err := c.DialOrDie().CoreV1().ConfigMaps(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ConfigMap.
func (c *ConfigMap) Delete(ns, n string) error {
	return c.DialOrDie().CoreV1().ConfigMaps(ns).Delete(n, nil)
}
