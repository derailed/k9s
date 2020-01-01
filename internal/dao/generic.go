package dao

import (
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/dynamic"
)

// Generic represents a generic resource.
type Generic struct {
	Factory

	gvr client.GVR
}

// Init initializes the resource.
func (g *Generic) Init(f Factory, gvr client.GVR) {
	g.Factory, g.gvr = f, gvr
}

// Delete a Generic.
func (g *Generic) Delete(path string, cascade, force bool) error {
	ns, n := client.Namespaced(path)
	auth, err := g.Client().CanI(ns, g.gvr.String(), []string{"delete"})
	if !auth || err != nil {
		return err
	}

	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	opts := metav1.DeleteOptions{PropagationPolicy: &p}
	if ns != "-" {
		return g.dynClient().Namespace(ns).Delete(n, &opts)
	}
	return g.dynClient().Delete(n, &opts)
}

func (g *Generic) dynClient() dynamic.NamespaceableResourceInterface {
	return g.Client().DynDialOrDie().Resource(g.gvr.AsGVR())
}
