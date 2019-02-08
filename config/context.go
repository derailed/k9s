package config

import (
	log "github.com/sirupsen/logrus"
)

// Context tracks K9s cluster context configuration.
type Context struct {
	Active   string              `yaml:"active"`
	Clusters map[string]*Cluster `yaml:"clusters"`
}

// NewContext creates a new cluster config context.
func NewContext() *Context {
	return &Context{Clusters: make(map[string]*Cluster, 1)}
}

// SetActiveCluster set the active cluster.
func (c *Context) SetActiveCluster(s string) {
	c.Active = s
	if _, ok := c.Clusters[c.Active]; ok {
		return
	}
	c.Clusters[c.Active] = NewCluster()
	return
}

// ActiveClusterName returns the currently active cluster name.
func (c *Context) ActiveClusterName() string {
	return c.Active
}

// ActiveCluster returns the currently active cluster configuration.
func (c *Context) ActiveCluster() *Cluster {
	if cl, ok := c.Clusters[c.Active]; ok {
		return cl
	}
	c.Clusters[c.Active] = NewCluster()
	return c.Clusters[c.Active]
}

// Validate this configuration
func (c *Context) Validate(ci ClusterInfo) {
	if len(c.Active) == 0 {
		c.Active = ci.ActiveClusterOrDie()
	}

	if c.Clusters == nil {
		c.Clusters = make(map[string]*Cluster, 1)
	}

	cc := ci.AllClustersOrDie()
	if len(cc) == 0 {
		log.Panic("Unable to find any live clusters in this configuration")
	}
	if !InList(cc, c.Active) {
		c.Active = cc[0]
		c.Clusters[cc[0]] = NewCluster()
	}

	if len(c.Clusters) == 0 {
		c.Clusters[c.Active] = NewCluster()
	}

	for k, cl := range c.Clusters {
		if !InList(cc, k) {
			delete(c.Clusters, k)
		} else {
			cl.Validate(ci)
		}
	}
}

func (c *Context) activeCluster() *Cluster {
	return c.Clusters[c.Active]
}
