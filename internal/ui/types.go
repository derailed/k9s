package ui

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

type (
	// SortFn represent a function that can sort columnar data.
	SortFn func(rows render.Rows, sortCol SortColumn)

	// SortColumn represents a sortable column.
	SortColumn struct {
		name string
		asc  bool
	}
)

// Namespaceable represents a namespaceable model.
type Namespaceable interface {
	// ClusterWide returns true if the model represents resource in all namespaces.
	ClusterWide() bool

	// GetNamespace returns the model namespace.
	GetNamespace() string

	// SetNamespace changes the model namespace.
	SetNamespace(string)

	// InNamespace check if current namespace matches models.
	InNamespace(string) bool
}

// Lister represents a viewable resource.
type Lister interface {
	// Get returns a resource instance.
	Get(ctx context.Context, path string) (runtime.Object, error)
}

// Tabular represents a tabular model.
type Tabular interface {
	Namespaceable
	Lister

	// SetInstance sets parent resource path.
	SetInstance(string)

	// SetLabelFilter sets the label filter.
	SetLabelFilter(string)

	// Empty returns true if model has no data.
	Empty() bool

	// Count returns the model data count.
	Count() int

	// Peek returns current model data.
	Peek() render.TableData

	// Watch watches a given resource for changes.
	Watch(context.Context) error

	// Refresh forces a new refresh.
	Refresh(context.Context) error

	// SetRefreshRate sets the model watch loop rate.
	SetRefreshRate(time.Duration)

	// AddListener registers a model listener.
	AddListener(model.TableListener)

	// RemoveListener unregister a model listener.
	RemoveListener(model.TableListener)

	// Delete a resource.
	Delete(ctx context.Context, path string, cascade, force bool) error
}
