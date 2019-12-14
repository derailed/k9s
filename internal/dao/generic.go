package dao

import (
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

type Generic struct {
	Factory

	gvr client.GVR
}

func (r *Generic) Init(f Factory, gvr client.GVR) {
	r.Factory, r.gvr = f, gvr
}

// Delete a Generic.
func (g *Generic) Delete(path string, cascade, force bool) error {
	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}

	ns, n := client.Namespaced(path)
	return g.dynClient().Namespace(ns).Delete(n, &metav1.DeleteOptions{
		PropagationPolicy: &p,
	})
}

func (g *Generic) dynClient() dynamic.NamespaceableResourceInterface {
	return g.Client().DynDialOrDie().Resource(g.gvr.AsGVR())
}
