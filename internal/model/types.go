// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tview"
	"github.com/sahilm/fuzzy"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	maxReaderRetryInterval   = 2 * time.Minute
	defaultReaderRefreshRate = 5 * time.Second
)

// ResourceViewerListener listens to viewing resource events.
type ResourceViewerListener interface {
	ResourceChanged(lines []string, matches fuzzy.Matches)
	ResourceFailed(error)
}

// ViewerToggleOpts represents a collection of viewing options.
type ViewerToggleOpts map[string]bool

// ResourceViewer represents a viewed resource.
type ResourceViewer interface {
	GetPath() string
	Filter(string)
	GVR() client.GVR
	ClearFilter()
	Peek() []string
	SetOptions(context.Context, ViewerToggleOpts)
	Watch(context.Context) error
	Refresh(context.Context) error
	AddListener(ResourceViewerListener)
	RemoveListener(ResourceViewerListener)
}

// EncDecResourceViewer interface extends the ResourceViewer interface and
// adds a `Toggle` that allows the user to switch between encoded or decoded
// state of the view.
type EncDecResourceViewer interface {
	ResourceViewer
	Toggle()
}

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

	// ExtraHints returns additional hints.
	ExtraHints() map[string]string
}

// Primitive represents a UI primitive.
type Primitive interface {
	tview.Primitive

	// Name returns the view name.
	Name() string
}

// Commander tracks prompt status.
type Commander interface {
	// InCmdMode checks if prompt is active.
	InCmdMode() bool
}

// Component represents a ui component.
type Component interface {
	Primitive
	Igniter
	Hinter
	Commander
	Filterer
}

type Filterer interface {
	SetFilter(string)
	SetLabelFilter(map[string]string)
}

// Cruder performs crud operations.
type Cruder interface {
	// List returns a collection of resources.
	List(ctx context.Context, ns string) ([]runtime.Object, error)

	// Get returns a resource instance.
	Get(ctx context.Context, path string) (runtime.Object, error)
}

// Lister represents a resource lister.
type Lister interface {
	Cruder

	// Init initializes a resource.
	Init(ns, gvr string, f dao.Factory)
}

// Describer represents a resource describer.
type Describer interface {
	// ToYAML return resource yaml.
	ToYAML(ctx context.Context, path string) (string, error)

	// Describe returns a resource description.
	Describe(client client.Connection, gvr, path string) (string, error)
}

// TreeRenderer represents an xray node.
type TreeRenderer interface {
	Render(ctx context.Context, ns string, o interface{}) error
}

// ResourceMeta represents model info about a resource.
type ResourceMeta struct {
	DAO          dao.Accessor
	Renderer     model1.Renderer
	TreeRenderer TreeRenderer
}
