package resource

import (
	"github.com/k8sland/k9s/resource/k8s"
	"k8s.io/api/core/v1"
)

type (
	// ClusterIfc represents a cluster.
	ClusterIfc interface {
		Version() (string, error)
		ClusterName() string
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

// Name returns the cluster name
func (c *Cluster) Name() string {
	return c.api.ClusterName()
}

// Metrics gathers node level metrics and compute utilization percentages.
func (c *Cluster) Metrics() (k8s.Metric, error) {
	return c.mx.NodeMetrics()
}
