package dao

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

var _ Describer = (*Generic)(nil)

// Generic represents a generic resource.
type Generic struct {
	NonResource
}

// Describe describes a resource.
func (g *Generic) Describe(path string) (string, error) {
	return Describe(g.Client(), g.gvr, path)
}

// ToYAML returns a resource yaml.
func (g *Generic) ToYAML(path string) (string, error) {
	o, err := g.Get(context.Background(), path)
	if err != nil {
		return "", err
	}

	raw, err := ToYAML(o)
	if err != nil {
		return "", fmt.Errorf("unable to marshal resource %s", err)
	}
	return raw, nil
}

// List returns a collection of resources.
func (g *Generic) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labelSel, ok := ctx.Value(internal.KeyLabels).(string)
	if !ok {
		log.Warn().Msgf("No label selector found in context. Listing all resources")
	}

	if client.IsAllNamespace(ns) {
		ns = client.AllNamespaces
	}

	var (
		ll  *unstructured.UnstructuredList
		err error
	)
	if client.IsNamespaced(ns) {
		ll, err = g.dynClient().Namespace(ns).List(metav1.ListOptions{LabelSelector: labelSel})
	} else {
		ll, err = g.dynClient().List(metav1.ListOptions{LabelSelector: labelSel})
	}
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, len(ll.Items))
	for i := range ll.Items {
		oo[i] = &ll.Items[i]
	}

	return oo, nil
}

// Get returns a given resource.
func (g *Generic) Get(ctx context.Context, path string) (runtime.Object, error) {
	var opts metav1.GetOptions

	ns, n := client.Namespaced(path)
	req := g.dynClient()
	if client.IsClusterScoped(ns) {
		return req.Get(n, opts)
	}

	return req.Namespace(ns).Get(n, opts)
}

// Delete deletes a resource.
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
	if client.IsClusterScoped(ns) {
		return g.dynClient().Delete(n, &opts)
	}

	return g.dynClient().Namespace(ns).Delete(n, &opts)
}

func (g *Generic) dynClient() dynamic.NamespaceableResourceInterface {
	return g.Client().DynDialOrDie().Resource(g.gvr.AsGVR())
}
