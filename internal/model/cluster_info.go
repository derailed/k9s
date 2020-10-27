package model

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
)

// ClusterInfoListener registers a listener for model changes.
type ClusterInfoListener interface {
	// ClusterInfoChanged notifies the cluster meta was changed.
	ClusterInfoChanged(prev, curr ClusterMeta)

	// ClusterInfoUpdated notifies the cluster meta was updated.
	ClusterInfoUpdated(ClusterMeta)
}

// ClusterMeta represents cluster meta data.
type ClusterMeta struct {
	Context, Cluster    string
	User                string
	K9sVer, K8sVer      string
	Cpu, Mem, Ephemeral int
}

// NewClusterMeta returns a new instance.
func NewClusterMeta() ClusterMeta {
	return ClusterMeta{
		Context:   client.NA,
		Cluster:   client.NA,
		User:      client.NA,
		K9sVer:    client.NA,
		K8sVer:    client.NA,
		Cpu:       0,
		Mem:       0,
		Ephemeral: 0,
	}
}

// Deltas diffs cluster meta return true if different, false otherwise.
func (c ClusterMeta) Deltas(n ClusterMeta) bool {
	if c.Cpu != n.Cpu {
		return true
	}
	if c.Mem != n.Mem {
		return true
	}
	if c.Ephemeral != n.Ephemeral {
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
	data := NewClusterMeta()
	data.Context = c.cluster.ContextName()
	data.Cluster = c.cluster.ClusterName()
	data.User = c.cluster.UserName()
	data.K9sVer = c.version
	data.K8sVer = c.cluster.Version()

	ctx, cancel := context.WithTimeout(context.Background(), c.cluster.factory.Client().Config().CallTimeout())
	defer cancel()
	var mx client.ClusterMetrics
	if err := c.cluster.Metrics(ctx, &mx); err == nil {
		data.Cpu, data.Mem, data.Ephemeral = mx.PercCPU, mx.PercMEM, mx.PercEphemeral
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
