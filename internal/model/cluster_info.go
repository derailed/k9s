package model

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
)

// ClusterInfoListener registers a listener for model changes.
type ClusterInfoListener interface {
	// ClusterInfoChanged notifies the cluster meta was changed.
	ClusterInfoChanged(prev, curr ClusterMeta)

	// ClusterInfoUpdated notifies the cluster meta was updated.
	ClusterInfoUpdated(ClusterMeta)
}

// NA indicates data is missing at this time.
const NA = "n/a"

// ClusterMeta represents cluster meta data.
type ClusterMeta struct {
	Context, Cluster string
	User             string
	K9sVer, K8sVer   string
	Cpu, Mem         float64
}

// NewClusterMeta returns a new instance.
func NewClusterMeta() ClusterMeta {
	return ClusterMeta{
		Context: NA,
		Cluster: NA,
		User:    NA,
		K9sVer:  NA,
		K8sVer:  NA,
		Cpu:     0,
		Mem:     0,
	}
}

// Deltas diffs cluster meta return true if different, false otherwise.
func (c ClusterMeta) Deltas(n ClusterMeta) bool {
	if render.AsPerc(c.Cpu) != render.AsPerc(n.Cpu) {
		return true
	}
	if render.AsPerc(c.Mem) != render.AsPerc(n.Mem) {
		return true
	}

	return c.Context != n.Context ||
		c.Cluster != n.Cluster ||
		c.User != n.User ||
		c.K8sVer != n.K8sVer ||
		c.K9sVer != n.K9sVer
}

// ClusterInfo models cluster metadata.
type ClusterInfo struct {
	cluster   *Cluster
	data      ClusterMeta
	version   string
	listeners []ClusterInfoListener
}

// NewClusterInfo returns a new instance.
func NewClusterInfo(f dao.Factory, version string) *ClusterInfo {
	return &ClusterInfo{
		cluster: NewCluster(f),
		data:    NewClusterMeta(),
		version: version,
	}
}

// Reset resets context and reload.
func (c *ClusterInfo) Reset(f dao.Factory) {
	c.cluster, c.data = NewCluster(f), NewClusterMeta()
	c.Refresh()
}

// Refresh fetches latest cluster meta.
func (c *ClusterInfo) Refresh() {
	log.Debug().Msgf("Refreshing ClusterInfo...")
	data := NewClusterMeta()
	data.Context = c.cluster.ContextName()
	data.Cluster = c.cluster.ClusterName()
	data.User = c.cluster.UserName()
	data.K9sVer = c.version
	data.K8sVer = c.cluster.Version()

	var mx client.ClusterMetrics
	if err := c.cluster.Metrics(&mx); err == nil {
		data.Cpu, data.Mem = mx.PercCPU, mx.PercMEM
	}

	if c.data.Deltas(data) {
		c.fireMetaChanged(c.data, data)
	} else {
		c.fireNoMetaChanged(data)
	}
	c.data = data
}

// AddListener adds a new model listener.
func (c *ClusterInfo) AddListener(l ClusterInfoListener) {
	c.listeners = append(c.listeners, l)
}

// RemoveListener delete a listener from the list.
func (c *ClusterInfo) RemoveListener(l ClusterInfoListener) {
	victim := -1
	for i, lis := range c.listeners {
		if lis == l {
			victim = i
			break
		}
	}

	if victim >= 0 {
		c.listeners = append(c.listeners[:victim], c.listeners[victim+1:]...)
	}
}

func (c *ClusterInfo) fireMetaChanged(prev, cur ClusterMeta) {
	for _, l := range c.listeners {
		l.ClusterInfoChanged(prev, cur)
	}
}

func (c *ClusterInfo) fireNoMetaChanged(data ClusterMeta) {
	for _, l := range c.listeners {
		l.ClusterInfoUpdated(data)
	}
}
