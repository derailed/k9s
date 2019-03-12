package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ConfigMap represents a Kubernetes ConfigMap
type ConfigMap struct{}

// NewConfigMap returns a new ConfigMap.
func NewConfigMap() Res {
	return &ConfigMap{}
}

// Get a ConfigMap.
func (c *ConfigMap) Get(ns, n string) (interface{}, error) {
	opts := metav1.GetOptions{}
	return conn.dialOrDie().CoreV1().ConfigMaps(ns).Get(n, opts)
}

// List all ConfigMaps in a given namespace.
func (c *ConfigMap) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{}

	rr, err := conn.dialOrDie().CoreV1().ConfigMaps(ns).List(opts)
	if err != nil {
		return Collection{}, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a ConfigMap.
func (c *ConfigMap) Delete(ns, n string) error {
	opts := metav1.DeleteOptions{}
	return conn.dialOrDie().CoreV1().ConfigMaps(ns).Delete(n, &opts)
}
