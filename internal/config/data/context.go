// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"sync"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/client-go/tools/clientcmd/api"
)

// DefaultPFAddress specifies the default PortForward host address.
const DefaultPFAddress = "localhost"

// Context tracks K9s context configuration.
type Context struct {
	ClusterName        string       `yaml:"cluster,omitempty"`
	ReadOnly           *bool        `yaml:"readOnly,omitempty"`
	Skin               string       `yaml:"skin,omitempty"`
	Namespace          *Namespace   `yaml:"namespace"`
	View               *View        `yaml:"view"`
	FeatureGates       FeatureGates `yaml:"featureGates"`
	PortForwardAddress string       `yaml:"portForwardAddress"`
	mx                 sync.RWMutex
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

// NewContextFromConfig returns a config based on a kubecontext.
func NewContextFromConfig(cfg *api.Context) *Context {
	return &Context{
		Namespace:          NewActiveNamespace(cfg.Namespace),
		ClusterName:        cfg.Cluster,
		View:               NewView(),
		PortForwardAddress: DefaultPFAddress,
		FeatureGates:       NewFeatureGates(),
	}
}

// NewContextFromKubeConfig returns a new instance based on kubesettings or an error.
func NewContextFromKubeConfig(ks KubeSettings) (*Context, error) {
	ct, err := ks.CurrentContext()
	if err != nil {
		return nil, err
	}

	return NewContextFromConfig(ct), nil
}

func (c *Context) GetClusterName() string {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.ClusterName
}

// Validate ensures a context config is tip top.
func (c *Context) Validate(conn client.Connection, ks KubeSettings) {
	c.mx.Lock()
	defer c.mx.Unlock()

	if c.PortForwardAddress == "" {
		c.PortForwardAddress = DefaultPFAddress
	}
	if cl, err := ks.CurrentClusterName(); err == nil {
		c.ClusterName = cl
	}

	if c.Namespace == nil {
		c.Namespace = NewNamespace()
	}
	c.Namespace.Validate(conn)

	if c.View == nil {
		c.View = NewView()
	}
	c.View.Validate()
}
