package dao

import (
	"context"
	"io"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/watch"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	restclient "k8s.io/client-go/rest"
)

// ResourceMetas represents a collection of resource metadata.
type ResourceMetas map[client.GVR]metav1.APIResource

// Accessors represents a collection of dao accessors.
type Accessors map[client.GVR]Accessor

// Factory represents a resource factory.
type Factory interface {
	// Client retrieves an api client.
	Client() client.Connection

	// Get fetch a given resource.
	Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error)

	// List fetch a collection of resources.
	List(gvr, ns string, wait bool, sel labels.Selector) ([]runtime.Object, error)

	// ForResource fetch an informer for a given resource.
	ForResource(ns, gvr string) informers.GenericInformer

	// CanForResource fetch an informer for a given resource if authorized
	CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error)

	// WaitForCacheSync synchronize the cache.
	WaitForCacheSync()

	// DeleteForwarder deletes a pod forwarder.
	DeleteForwarder(path string)

	// Forwards returns all portforwards.
	Forwarders() watch.Forwarders
}

// Getter represents a resource getter.
type Getter interface {
	// Get return a given resource.
	Get(ctx context.Context, path string) (runtime.Object, error)
}

// Lister represents a resource lister.
type Lister interface {
	// List returns a resource collection.
	List(ctx context.Context, ns string) ([]runtime.Object, error)
}

// Accessor represents an accessible k8s resource.
type Accessor interface {
	Lister
	Getter

	// Init the resource with a factory object.
	Init(Factory, client.GVR)

	// GVR returns a gvr a string.
	GVR() string
}

// DrainOptions tracks drain attributes.
type DrainOptions struct {
	GracePeriodSeconds  int
	Timeout             time.Duration
	IgnoreAllDaemonSets bool
	DeleteLocalData     bool
	Force               bool
}

// NodeMaintainer performs node maintenance operations.
type NodeMaintainer interface {
	// ToggleCordon toggles cordon/uncordon a node.
	ToggleCordon(path string, cordon bool) error

	// Drain drains the given node.
	Drain(path string, opts DrainOptions, w io.Writer) error
}

// Loggable represents resources with logs.
type Loggable interface {
	// TaiLogs streams resource logs.
	TailLogs(ctx context.Context, c LogChan, opts LogOptions) error
}

// Describer describes a resource.
type Describer interface {
	// Describe describes a resource.
	Describe(path string) (string, error)

	// ToYAML dumps a resource to YAML.
	ToYAML(path string) (string, error)
}

// Scalable represents resources that can scale.
type Scalable interface {
	// Scale scales a resource up or down.
	Scale(ctx context.Context, path string, replicas int32) error
}

// Controller represents a pod controller.
type Controller interface {
	// Pod returns a pod instance matching the selector.
	Pod(path string) (string, error)
}

// Nuker represents a resource deleter.
type Nuker interface {
	// Delete removes a resource from the api server.
	Delete(path string, cascade, force bool) error
}

// Switchable represents a switchable resource.
type Switchable interface {
	// Switch changes the active context.
	Switch(ctx string) error
}

// Restartable represents a restartable resource.
type Restartable interface {
	// Restart performs a rollout restart.
	Restart(ctx context.Context, path string) error
}

// Runnable represents a runnable resource.
type Runnable interface {
	// Run triggers a run.
	Run(path string) error
}

// Logger represents a resource that exposes logs.
type Logger interface {
	// Logs tails a resource logs.
	Logs(path string, opts *v1.PodLogOptions) (*restclient.Request, error)
}
