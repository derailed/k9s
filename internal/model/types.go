package model

import (
	"context"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/informers"
)

// Igniter represents a runnable view.
type Igniter interface {
	// Start starts a component.
	Init(ctx context.Context) error

	// Start starts a component.
	Start()

	// Stop terminates a component.
	Stop()
}

// Hinter represent a menu mnemonic provider.
type Hinter interface {
	// Hints returns a collection of menu hints.
	Hints() MenuHints
}

// Primitive represents a UI primitive.
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

// Component represents a ui component
type Component interface {
	Primitive
	Igniter
	Hinter
}

// Renderer represents a resource renderer.
type Renderer interface {
	// Render converts raw resources to tabular data.
	Render(o interface{}, ns string, row *render.Row) error

	// Header returns the resource header.
	Header(ns string) render.HeaderRow

	// ColorerFunc returns a row colorer function.
	ColorerFunc() render.ColorerFunc
}

// Lister represents a resource lister.
type Lister interface {
	// Init initializes a resource.
	Init(ns, gvr string, f Factory)

	// List returns a collection of resources.
	List(context.Context) ([]runtime.Object, error)

	// Hydrate converts resource rows into tabular data.
	Hydrate([]runtime.Object, render.Rows, Renderer) error
}

// BOZO!!
// type Connection interface {
// 	// DialOrDie dials client api.
// 	DialOrDie() kubernetes.Interface

// 	// MXDial dials metrics api.
// 	MXDial() (*versioned.Clientset, error)

// 	// DynDialOrDie dials dynamic client api.
// 	DynDialOrDie() dynamic.Interface

// 	// RestConfigOrDie return a client configuration.
// 	RestConfigOrDie() *restclient.Config

// 	// Config returns the current kubeconfig.
// 	Config() *k8s.Config

// 	// CachedDiscovery returns a cached client.
// 	CachedDiscovery() (*disk.CachedDiscoveryClient, error)

// 	// SwithContextOrDie switch to a new kube context.
// 	SwitchContextOrDie(ctx string)
// }

type Factory interface {
	// Client retrieves an api client.
	Client() k8s.Connection

	// Get fetch a given resource.
	Get(ns, gvr, n string, sel labels.Selector) (runtime.Object, error)

	// List fetch a collection of resources.
	List(ns, gvr string, sel labels.Selector) ([]runtime.Object, error)

	// ForResource fetch an informer for a given resource.
	ForResource(ns, gvr string) informers.GenericInformer

	// WaitForCacheSync synchronize the cache.
	WaitForCacheSync() map[schema.GroupVersionResource]bool
}

// ResourceMeta represents model info about a resource.
type ResourceMeta struct {
	Model    Lister
	Renderer Renderer
}
