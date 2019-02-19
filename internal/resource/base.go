package resource

import (
	"path"

	"github.com/derailed/k9s/internal/k8s"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// Creator can create a new resources.
	Creator interface {
		NewInstance(interface{}) Columnar
	}

	// Caller can call Kubernetes verbs on a resource.
	Caller interface {
		k8s.Res
	}

	// APIFn knows how to call K8s api server.
	APIFn func() k8s.Res

	// InstanceFn instantiates a concrete resource.
	InstanceFn func(interface{}) Columnar

	// Base resource.
	Base struct {
		path string

		caller  Caller
		creator Creator
	}
)

// Name returns the resource namespaced name.
func (b *Base) Name() string {
	return b.path
}

// Get a resource by name
func (b *Base) Get(path string) (Columnar, error) {
	ns, n := namespaced(path)
	i, err := b.caller.Get(ns, n)
	if err != nil {
		return nil, err
	}
	return b.creator.NewInstance(i), nil
}

// List all resources
func (b *Base) List(ns string) (Columnars, error) {
	ii, err := b.caller.List(ns)
	if err != nil {
		return nil, err
	}

	cc := make(Columnars, 0, len(ii))
	for i := 0; i < len(ii); i++ {
		cc = append(cc, b.creator.NewInstance(ii[i]))
	}
	return cc, nil
}

// Delete a resource by name.
func (b *Base) Delete(path string) error {
	ns, n := namespaced(path)
	return b.caller.Delete(ns, n)
}

func (*Base) namespacedName(m metav1.ObjectMeta) string {
	return path.Join(m.Namespace, m.Name)
}
