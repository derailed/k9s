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
)

const (
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

	Config() *Config
	DialOrDie() kubernetes.Interface
	SwitchContextOrDie(ctx string)
	CachedDiscoveryOrDie() *disk.CachedDiscoveryClient
	RestConfigOrDie() *restclient.Config
	MXDial() (*versioned.Clientset, error)
	DynDialOrDie() dynamic.Interface
	HasMetrics() bool
	ValidNamespaces() ([]v1.Namespace, error)
	ServerVersion() (*version.Info, error)
	CheckConnectivity() bool
}

// CurrentMetrics tracks current cpu/mem.
type CurrentMetrics struct {
	CurrentCPU int64
	CurrentMEM float64
}

// PodMetrics represent an aggregation of all pod containers metrics.
type PodMetrics CurrentMetrics

// NodeMetrics describes raw node metrics.
type NodeMetrics struct {
	CurrentMetrics
	AvailCPU int64
	AvailMEM float64
	TotalCPU int64
	TotalMEM float64
}

// ClusterMetrics summarizes total node metrics as percentages.
type ClusterMetrics struct {
	PercCPU float64
	PercMEM float64
}

// NodesMetrics tracks usage metrics per nodes.
type NodesMetrics map[string]NodeMetrics

// PodsMetrics tracks usage metrics per pods.
type PodsMetrics map[string]PodMetrics
