package model

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
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
	Hydrate(oo []runtime.Object, rr render.Rows, r Renderer) error
}

type Factory interface {
	// Client retrieves an api client.
	Client() client.Connection

	// Get fetch a given resource.
	Get(gvr, path string, sel labels.Selector) (runtime.Object, error)

	// List fetch a collection of resources.
	List(gvr, ns string, sel labels.Selector) ([]runtime.Object, error)

	// ForResource fetch an informer for a given resource.
	ForResource(ns, gvr string) informers.GenericInformer

	// CanForResource fetch an informer for a given resource.
	CanForResource(ns, gvr string, verbs ...string) (informers.GenericInformer, error)

	// WaitForCacheSync synchronize the cache.
	WaitForCacheSync()

	// Forwards returns all portforwards.
	Forwarders() watch.Forwarders
}

// ResourceMeta represents model info about a resource.
type ResourceMeta struct {
	Model    Lister
	Renderer Renderer
}
