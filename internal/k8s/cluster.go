package k8s

import (
	"github.com/rs/zerolog"
	v1 "k8s.io/api/core/v1"
)

const na = "n/a"

// Cluster represents a Kubernetes cluster.
type Cluster struct {
	Connection

	logger *zerolog.Logger
}

// NewCluster instantiates a new cluster.
func NewCluster(c Connection, l *zerolog.Logger) *Cluster {
	return &Cluster{c, l}
}

// Version returns the current cluster git version.
func (c *Cluster) Version() (string, error) {
	rev, err := c.ServerVersion()
	if err != nil {
		c.logger.Warn().Msgf("%s", err)
		return "", err
	}
	return rev.GitVersion, nil
}

// ContextName returns the currently active context.
func (c *Cluster) ContextName() string {
	ctx, err := c.Config().CurrentContextName()
	if err != nil {
		c.logger.Warn().Msgf("%s", err)
		return na
	}
	return ctx
}

// ClusterName return the currently active cluster name.
func (c *Cluster) ClusterName() string {
	ctx, err := c.Config().CurrentClusterName()
	if err != nil {
		c.logger.Warn().Msgf("%s", err)
		return na
	}
	return ctx
}

// UserName returns the currently active user.
func (c *Cluster) UserName() string {
	usr, err := c.Config().CurrentUserName()
	if err != nil {
		c.logger.Warn().Msgf("%s", err)
		return na
	}
	return usr
}

// GetNodes get all available nodes in the cluster.
func (c *Cluster) GetNodes() (*v1.NodeList, error) {
	return c.FetchNodes()
}
