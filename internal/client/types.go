package client

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// NA Not available
	NA = "n/a"

	// NamespaceAll designates the fictional all namespace.
	NamespaceAll = "all"

	// AllNamespaces designates all namespaces.
	AllNamespaces = ""

	// ClusterScope designates a resource is not namespaced.
	ClusterScope = "-"

	// NotNamespaced designates a non resource namespace.
	NotNamespaced = "*"

	// CreateVerb represents create access on a resource.
	CreateVerb = "create"

	// UpdateVerb represents an update access on a resource.
	UpdateVerb = "update"

	// PatchVerb represents a patch access on a resource.
	PatchVerb = "patch"

	// DeleteVerb represents a delete access on a resource.
	DeleteVerb = "delete"

	// GetVerb represents a get access on a resource.
	GetVerb = "get"

	// ListVerb represents a list access on a resource.
	ListVerb = "list"

	// WatchVerb represents a watch access on a resource.
	WatchVerb = "watch"
)

var (
	// GetAccess reads a resource.
	GetAccess = []string{GetVerb}
	// ListAccess list resources.
	ListAccess = []string{ListVerb}
	// MonitorAccess monitors a collection of resources.
	MonitorAccess = []string{ListVerb, WatchVerb}
	// ReadAllAccess represents an all read access to a resource.
	ReadAllAccess = []string{GetVerb, ListVerb, WatchVerb}
)

// Authorizer checks what a user can or cannot do to a resource.
type Authorizer interface {
	// CanI returns true if the user can use these actions for a given resource.
	CanI(ns, gvr string, verbs []string) (bool, error)
}

// Connection represents a Kubenetes apiserver connection.
type Connection interface {
	Authorizer

	// Config returns current config.
	Config() *Config

	// DialOrDie connects to api server.
	DialOrDie() kubernetes.Interface

	// SwitchContext switches cluster based on context.
	SwitchContext(ctx string) error

	// CachedDiscoveryOrDie connects to discovery client.
	CachedDiscoveryOrDie() *disk.CachedDiscoveryClient

	// RestConfigOrDie connects to rest client.
	RestConfigOrDie() *restclient.Config

	// MXDial connects to metrics server.
	MXDial() (*versioned.Clientset, error)

	// DynDialOrDie connects to dynamic client.
	DynDialOrDie() dynamic.Interface

	// HasMetrics checks if metrics server is available.
	HasMetrics() bool

	// ValidNamespaces returns all available namespaces.
	ValidNamespaces() ([]v1.Namespace, error)

	// ServerVersion returns current server version.
	ServerVersion() (*version.Info, error)

	// CheckConnectivity checks if api server connection is happy or not.
	CheckConnectivity() bool

	// ActiveCluster returns the current cluster name.
	ActiveCluster() string

	// ActiveNamespace returns the current namespace.
	ActiveNamespace() string

	// IsActiveNamespace checks if given ns is active.
	IsActiveNamespace(string) bool
}

// CurrentMetrics tracks current cpu/mem.
type CurrentMetrics struct {
	CurrentCPU, CurrentMEM, CurrentEphemeral int64
}

// PodMetrics represent an aggregation of all pod containers metrics.
type PodMetrics CurrentMetrics

// NodeMetrics describes raw node metrics.
type NodeMetrics struct {
	CurrentMetrics

	AllocatableCPU, AllocatableMEM, AllocatableEphemeral int64
	AvailableCPU, AvailableMEM, AvailableEphemeral       int64
	TotalCPU, TotalMEM, TotalEphemeral                   int64
}

// ClusterMetrics summarizes total node metrics as percentages.
type ClusterMetrics struct {
	PercCPU, PercMEM, PercEphemeral int
}

// NodesMetrics tracks usage metrics per nodes.
type NodesMetrics map[string]NodeMetrics

// PodsMetrics tracks usage metrics per pods.
type PodsMetrics map[string]PodMetrics
