package dao

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type Resource struct {
	Factory

	gvr GVR
}

func (r *Resource) Init(f Factory, gvr GVR) {
	r.Factory, r.gvr = f, gvr
}

// Delete a Generic.
func (r *Resource) Delete(ns, n string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}

	return r.dynClient().Namespace(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}

func (r *Resource) dynClient() dynamic.NamespaceableResourceInterface {
	return r.Client().DynDialOrDie().Resource(r.gvr.AsGVR())
}
