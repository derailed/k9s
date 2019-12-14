package dao

import (
	"github.com/derailed/k9s/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type Generic struct {
	Factory

	gvr GVR
}

func (r *Generic) Init(f Factory, gvr GVR) {
	r.Factory, r.gvr = f, gvr
}

// Delete a Generic.
func (g *Generic) Delete(path string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}

	ns, n := k8s.Namespaced(path)
	return g.dynClient().Namespace(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}

func (g *Generic) dynClient() dynamic.NamespaceableResourceInterface {
	return g.Client().DynDialOrDie().Resource(g.gvr.AsGVR())
}
