package dao

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/watch"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/informers"
	restclient "k8s.io/client-go/rest"
)

type Factory interface {
	// Client retrieves an api client.
	Client() client.Connection

	// Get fetch a given resource.
	Get(gvr, path string, sel labels.Selector) (runtime.Object, error)

	// List fetch a collection of resources.
	List(gvr, ns string, sel labels.Selector) ([]runtime.Object, error)

	// ForResource fetch an informer for a given resource.
	ForResource(ns, gvr string) informers.GenericInformer

	// WaitForCacheSync synchronize the cache.
	WaitForCacheSync()

	// DeleteForwarder deletes a pod forwarder.
	DeleteForwarder(path string)

	// Forwards returns all portforwards.
	Forwarders() watch.Forwarders
}

// Accessor represents an accessible k8s resource.
type Accessor interface {
	Nuker

	// Init the resource with a factory object.
	Init(Factory, client.GVR)
}

// Loggable represents resources with logs.
type Loggable interface {
	// TaiLogs streams resource logs.
	TailLogs(ctx context.Context, c chan<- string, opts LogOptions) error
}

type Scalable interface {
	Scale(path string, replicas int32) error
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
	Restart(path string) error
}

// Runnable represents a runnable resource.
type Runnable interface {
	// Run triggers a run.
	Run(path string) error
}

// Loggers represents a resource that exposes logs.
type Logger interface {
	Logs(path string, opts *v1.PodLogOptions) *restclient.Request
}
