package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
)

var _ Describer = (*Generic)(nil)

// Generic represents a generic resource.
type Generic struct {
	NonResource
}

// List returns a collection of resources.
// BOZO!! no auth check??
func (g *Generic) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labelSel, _ := ctx.Value(internal.KeyLabels).(string)
	if client.IsAllNamespace(ns) {
		ns = client.AllNamespaces
	}

	var (
		ll  *unstructured.UnstructuredList
		err error
	)
	dial, err := g.dynClient()
	if err != nil {
		return nil, err
	}

	if client.IsClusterScoped(ns) {
		ll, err = dial.List(ctx, metav1.ListOptions{LabelSelector: labelSel})
	} else {
		ll, err = dial.Namespace(ns).List(ctx, metav1.ListOptions{LabelSelector: labelSel})
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
	dial, err := g.dynClient()
	if err != nil {
		return nil, err
	}
	if client.IsClusterScoped(ns) {
		return dial.Get(ctx, n, opts)
	}

	return dial.Namespace(ns).Get(ctx, n, opts)
}

// Describe describes a resource.
func (g *Generic) Describe(path string) (string, error) {
	return Describe(g.Client(), g.gvr, path)
}

// ToYAML returns a resource yaml.
func (g *Generic) ToYAML(path string, showManaged bool) (string, error) {
	o, err := g.Get(context.Background(), path)
	if err != nil {
		return "", err
	}

	raw, err := ToYAML(o, showManaged)
	if err != nil {
		return "", fmt.Errorf("unable to marshal resource %w", err)
	}
	return raw, nil
}

// Delete deletes a resource.
func (g *Generic) Delete(ctx context.Context, path string, propagation *metav1.DeletionPropagation, force bool) error {
	ns, n := client.Namespaced(path)
	auth, err := g.Client().CanI(ns, g.gvr.String(), []string{client.DeleteVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to delete %s", path)
	}

	opts := metav1.DeleteOptions{
		PropagationPolicy: propagation,
	}

	if force {
		var defaultKillGrace int64 = 1
		opts.GracePeriodSeconds = &defaultKillGrace
	}

	dial, err := g.dynClient()
	if err != nil {
		return err
	}
	if client.IsClusterScoped(ns) {
		return dial.Delete(ctx, n, opts)
	}
	ctx, cancel := context.WithTimeout(ctx, g.Client().Config().CallTimeout())
	defer cancel()

	return dial.Namespace(ns).Delete(ctx, n, opts)
}

func (g *Generic) dynClient() (dynamic.NamespaceableResourceInterface, error) {
	dial, err := g.Client().DynDial()
	if err != nil {
		return nil, err
	}

	return dial.Resource(g.gvr.GVR()), nil
}
