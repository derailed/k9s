package k8s

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
)

// Ingress represents a Kubernetes Ingress.
type Ingress struct {
	*Resource
	Connection
}

// NewIngress returns a new Ingress.
func NewIngress(c Connection, gvr GVR) *Ingress {
	return &Ingress{&Resource{gvr: gvr}, c}
}

func (i *Ingress) nsRes() dynamic.NamespaceableResourceInterface {
	g := schema.GroupVersionResource{
		Group:    i.gvr.Group,
		Version:  i.gvr.Version,
		Resource: i.gvr.Resource,
	}
	return i.DynDialOrDie().Resource(g)
}

// Get a Ingress.
func (i *Ingress) Get(ns, n string) (interface{}, error) {
	return i.nsRes().Namespace(ns).Get(n, metav1.GetOptions{})
}

// List all Ingresses in a given namespace.
func (i *Ingress) List(ns string) (Collection, error) {
	opts := metav1.ListOptions{
		LabelSelector: i.labelSelector,
		FieldSelector: i.fieldSelector,
	}
	rr, err := i.DialOrDie().ExtensionsV1beta1().Ingresses(ns).List(opts)
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
func (i *Ingress) Delete(ns, n string, cascade, force bool) error {
	return i.DialOrDie().ExtensionsV1beta1().Ingresses(ns).Delete(n, nil)
}
