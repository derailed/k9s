package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// StorageClass represents a Kubernetes StorageClass.
type StorageClass struct {
	*base
	Connection
}

// NewStorageClass returns a new StorageClass.
func NewStorageClass(c Connection) *StorageClass {
	return &StorageClass{&base{}, c}
}

// Get a StorageClass.
func (p *StorageClass) Get(_, n string) (interface{}, error) {
	return p.DialOrDie().StorageV1().StorageClasses().Get(n, metav1.GetOptions{})
}

// List all StorageClasses in a given namespace.
func (p *StorageClass) List(_ string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: p.labelSelector,
		FieldSelector: p.fieldSelector,
	}
	rr, err := p.DialOrDie().StorageV1().StorageClasses().List(opts)
	if err != nil {
		return nil, err
	}

	cc := make(Collection, len(rr.Items))
	for i, r := range rr.Items {
		cc[i] = r
	}

	return cc, nil
}

// Delete a StorageClass.
func (p *StorageClass) Delete(_, n string, cascade, force bool) error {
	return p.DialOrDie().StorageV1().StorageClasses().Delete(n, nil)
}
