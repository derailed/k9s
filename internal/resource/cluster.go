package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// ClusterMeta represents metadata about a Kubernetes cluster.
	ClusterMeta interface {
		Connection

		Version() (string, error)
		ContextName() string
		ClusterName() string
		UserName() string
		GetNodes() (*v1.NodeList, error)
	}

	// MetricsServer gather metrics information from pods and nodes.
	MetricsServer interface {
		MetricsService

		ClusterLoad(nodes k8s.Collection, metrics k8s.Collection, cmx *k8s.ClusterMetrics)
		NodesMetrics(k8s.Collection, *mv1beta1.NodeMetricsList, k8s.NodesMetrics)
		PodsMetrics(*mv1beta1.PodMetricsList, k8s.PodsMetrics)
	}

	// MetricsService calls the metrics server for metrics info.
	MetricsService interface {
		HasMetrics() bool
		FetchNodesMetrics() (*mv1beta1.NodeMetricsList, error)
		FetchPodsMetrics(ns string) (*mv1beta1.PodMetricsList, error)
	}

	// Cluster represents a kubernetes resource.
	Cluster struct {
		api ClusterMeta
		mx  MetricsServer
	}
)

// NewCluster returns a new cluster info resource.
func NewCluster(c Connection, log *zerolog.Logger, mx MetricsServer) *Cluster {
	return NewClusterWithArgs(k8s.NewCluster(c, log), mx)
}

// NewClusterWithArgs for tests only!
func NewClusterWithArgs(ci ClusterMeta, mx MetricsServer) *Cluster {
	return &Cluster{api: ci, mx: mx}
}

// Version returns the current K8s cluster version.
func (c *Cluster) Version() string {
	info, err := c.api.Version()
	if err != nil {
		return "n/a"
	}

	return info
}

// ContextName returns the context name.
func (c *Cluster) ContextName() string {
	return c.api.ContextName()
}

// ClusterName returns the cluster name.
func (c *Cluster) ClusterName() string {
	return c.api.ClusterName()
}

// UserName returns the user name.
func (c *Cluster) UserName() string {
	return c.api.UserName()
}

// Metrics gathers node level metrics and compute utilization percentages.
func (c *Cluster) Metrics(nos k8s.Collection, nmx k8s.Collection, mx *k8s.ClusterMetrics) {
	c.mx.ClusterLoad(nos, nmx, mx)
}
