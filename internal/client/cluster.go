package client

import (
	v1 "k8s.io/api/core/v1"
)

// Cluster represents a Kubernetes cluster.
type Cluster struct {
	Connection
}

// NewCluster instantiates a new cluster.
func NewCluster(c Connection) *Cluster {
	return &Cluster{Connection: c}
}

// Version returns the current cluster git version.
func (c *Cluster) Version() (string, error) {
	rev, err := c.ServerVersion()
	if err != nil {
		return "", err
	}
	return rev.GitVersion, nil
}

// ContextName returns the currently active context.
func (c *Cluster) ContextName() (string, error) {
	return c.Config().CurrentContextName()
}

// ClusterName return the currently active cluster name.
func (c *Cluster) ClusterName() (string, error) {
	return c.Config().CurrentClusterName()
}

// UserName returns the currently active user.
func (c *Cluster) UserName() (string, error) {
	return c.Config().CurrentUserName()
}

// GetNodes get all available nodes in the cluster.
func (c *Cluster) GetNodes() (*v1.NodeList, error) {
	return c.FetchNodes()
}
