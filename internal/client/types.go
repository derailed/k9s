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

// Connection represents a Kubenetes apiserver connection.
// BOZO!! Refactor!
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
	IsNamespaced(n string) bool
	SupportsResource(group string) bool
	ValidNamespaces() ([]v1.Namespace, error)
	SupportsRes(grp string, versions []string) (string, bool, error)
	ServerVersion() (*version.Info, error)
	CurrentNamespaceName() (string, error)
}

type currentMetrics struct {
	CurrentCPU int64
	CurrentMEM float64
}

// PodMetrics represent an aggregation of all pod containers metrics.
type PodMetrics currentMetrics

// NodeMetrics describes raw node metrics.
type NodeMetrics struct {
	currentMetrics
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
