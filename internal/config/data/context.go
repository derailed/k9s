// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"github.com/derailed/k9s/internal/client"
	"k8s.io/client-go/tools/clientcmd/api"
)

// DefaultPFAddress specifies the default PortForward host address.
const DefaultPFAddress = "localhost"

// Context tracks K9s context configuration.
type Context struct {
	ClusterName        string       `yaml:"cluster,omitempty"`
	ReadOnly           bool         `yaml:"readOnly"`
	Skin               string       `yaml:"skin,omitempty"`
	Namespace          *Namespace   `yaml:"namespace"`
	View               *View        `yaml:"view"`
	FeatureGates       FeatureGates `yaml:"featureGates"`
	PortForwardAddress string       `yaml:"portForwardAddress"`
}

// NewContext creates a new cluster configuration.
func NewContext() *Context {
	return &Context{
		Namespace:          NewNamespace(),
		View:               NewView(),
		PortForwardAddress: DefaultPFAddress,
		FeatureGates:       NewFeatureGates(),
	}
}

func NewContextFromConfig(cfg *api.Context) *Context {
	return &Context{
		Namespace:          NewActiveNamespace(cfg.Namespace),
		ClusterName:        cfg.Cluster,
		View:               NewView(),
		PortForwardAddress: DefaultPFAddress,
		FeatureGates:       NewFeatureGates(),
	}
}

// Validate a context config.
func (c *Context) Validate(conn client.Connection, ks KubeSettings) {
	if c.PortForwardAddress == "" {
		c.PortForwardAddress = DefaultPFAddress
	}

	if cl, err := ks.CurrentClusterName(); err != nil {
		c.ClusterName = cl
	}

	if c.Namespace == nil {
		c.Namespace = NewNamespace()
	}
	if c.Namespace.Active == client.BlankNamespace {
		c.Namespace.Active = client.DefaultNamespace
	}
	c.Namespace.Validate(conn, ks)

	if c.View == nil {
		c.View = NewView()
	}
	c.View.Validate()
}
