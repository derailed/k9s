// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// NA Not available.
	NA = "n/a"

	// NamespaceAll designates the fictional all namespace.
	NamespaceAll = "all"

	// BlankNamespace designates no namespace.
	BlankNamespace = ""

	// DefaultNamespace designates the default namespace.
	DefaultNamespace = "default"

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
	// PatchAccess patch a resource.
	PatchAccess = []string{PatchVerb}

	// GetAccess reads a resource.
	GetAccess = []string{GetVerb}

	// ListAccess list resources.
	ListAccess = []string{ListVerb}

	// MonitorAccess monitors a collection of resources.
	MonitorAccess = []string{ListVerb, WatchVerb}

	// ReadAllAccess represents an all read access to a resource.
	ReadAllAccess = []string{GetVerb, ListVerb, WatchVerb}
)

// ContainersMetrics tracks containers metrics.
type ContainersMetrics map[string]*mv1beta1.ContainerMetrics

// NodesMetricsMap tracks node metrics.
type NodesMetricsMap map[string]*mv1beta1.NodeMetrics

// PodsMetricsMap tracks pod metrics.
type PodsMetricsMap map[string]*mv1beta1.PodMetrics

// Authorizer checks what a user can or cannot do to a resource.
type Authorizer interface {
	// CanI returns true if the user can use these actions for a given resource.
	CanI(ns, gvr, n string, verbs []string) (bool, error)
}

// Connection represents a Kubernetes apiserver connection.
type Connection interface {
	Authorizer

	// Config returns current config.
	Config() *Config

	// ConnectionOK checks api server connection status.
	ConnectionOK() bool

	// Dial connects to api server.
	Dial() (kubernetes.Interface, error)

	// DialLogs connects to api server for logs.
	DialLogs() (kubernetes.Interface, error)

	// SwitchContext switches cluster based on context.
	SwitchContext(ctx string) error

	// CachedDiscovery connects to discovery client.
	CachedDiscovery() (*disk.CachedDiscoveryClient, error)

	// RestConfig connects to rest client.
	RestConfig() (*restclient.Config, error)

	// MXDial connects to metrics server.
	MXDial() (*versioned.Clientset, error)

	// DynDial connects to dynamic client.
	DynDial() (dynamic.Interface, error)

	// HasMetrics checks if metrics server is available.
	HasMetrics() bool

	// ValidNamespaceNames returns all available namespace names.
	ValidNamespaceNames() (NamespaceNames, error)

	// IsValidNamespace checks if given namespace is known.
	IsValidNamespace(string) bool

	// ServerVersion returns current server version.
	ServerVersion() (*version.Info, error)

	// CheckConnectivity checks if api server connection is happy or not.
	CheckConnectivity() bool

	// ActiveContext returns the current context name.
	ActiveContext() string

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
