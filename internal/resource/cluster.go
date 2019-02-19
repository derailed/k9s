package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	v1 "k8s.io/api/core/v1"
)

type (
	// ClusterIfc represents a cluster.
	ClusterIfc interface {
		Version() (string, error)
		ContextName() string
		ClusterName() string
		UserName() string
	}

	// MetricsIfc represents a metrics server.
	MetricsIfc interface {
		NodeMetrics() (k8s.Metric, error)
		PerNodeMetrics([]v1.Node) (map[string]k8s.Metric, error)
		PodMetrics() (map[string]k8s.Metric, error)
	}

	// Cluster represents a kubernetes resource.
	Cluster struct {
		api ClusterIfc
		mx  MetricsIfc
	}
)

// NewCluster returns a new cluster info resource.
func NewCluster() *Cluster {
	return NewClusterWithArgs(k8s.NewCluster(), k8s.NewMetricsServer())
}

// NewClusterWithArgs for tests only!
func NewClusterWithArgs(ci ClusterIfc, mx MetricsIfc) *Cluster {
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
func (c *Cluster) Metrics() (k8s.Metric, error) {
	return c.mx.NodeMetrics()
}
