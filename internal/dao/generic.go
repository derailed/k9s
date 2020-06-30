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

var defaultKillGrace int64

// Generic represents a generic resource.
type Generic struct {
	NonResource
}

// List returns a collection of resources.
// BOZO!! no auth check??
func (g *Generic) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	labelSel, ok := ctx.Value(internal.KeyLabels).(string)
	if !ok {
		log.Debug().Msgf("No label selector found in context. Listing all resources")
	}
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

// Delete deletes a resource.
func (g *Generic) Delete(path string, cascade, force bool) error {
	log.Debug().Msgf("DELETE %q -- %t:%t", path, cascade, force)
	ns, n := client.Namespaced(path)
	auth, err := g.Client().CanI(ns, g.gvr.String(), []string{client.DeleteVerb})
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf("user is not authorized to delete %s", path)
	}

	p := metav1.DeletePropagationOrphan
	if cascade {
		p = metav1.DeletePropagationBackground
	}
	var grace *int64
	if force {
		grace = &defaultKillGrace
	}
	opts := metav1.DeleteOptions{
		PropagationPolicy:  &p,
		GracePeriodSeconds: grace,
	}

	dial, err := g.dynClient()
	if err != nil {
		return err
	}
	// BOZO!! Move to caller!
	ctx, cancel := context.WithTimeout(context.Background(), g.Client().Config().CallTimeout())
	defer cancel()

	if client.IsClusterScoped(ns) {
		return dial.Delete(ctx, n, opts)
	}

	return dial.Namespace(ns).Delete(ctx, n, opts)
}

func (g *Generic) dynClient() (dynamic.NamespaceableResourceInterface, error) {
	dial, err := g.Client().DynDial()
	if err != nil {
		return nil, err
	}

	return dial.Resource(g.gvr.GVR()), nil
}
