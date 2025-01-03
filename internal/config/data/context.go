// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"os"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Context tracks K9s context configuration.
type Context struct {
	ClusterName        string       `yaml:"cluster,omitempty"`
	ReadOnly           *bool        `yaml:"readOnly,omitempty"`
	Skin               string       `yaml:"skin,omitempty"`
	Namespace          *Namespace   `yaml:"namespace"`
	View               *View        `yaml:"view"`
	FeatureGates       FeatureGates `yaml:"featureGates"`
	PortForwardAddress string       `yaml:"portForwardAddress"`
	Proxy              *Proxy       `yaml:"proxy"`
	mx                 sync.RWMutex
}

// NewContext creates a new cluster configuration.
func NewContext() *Context {
	return &Context{
		Namespace:          NewNamespace(),
		View:               NewView(),
		PortForwardAddress: defaultPFAddress(),
		FeatureGates:       NewFeatureGates(),
	}
}

// NewContextFromConfig returns a config based on a kubecontext.
func NewContextFromConfig(cfg *api.Context) *Context {
	ct := NewContext()
	ct.Namespace, ct.ClusterName = NewActiveNamespace(cfg.Namespace), cfg.Cluster

	return ct

}

// NewContextFromKubeConfig returns a new instance based on kubesettings or an error.
func NewContextFromKubeConfig(ks KubeSettings) (*Context, error) {
	ct, err := ks.CurrentContext()
	if err != nil {
		return nil, err
	}

	return NewContextFromConfig(ct), nil
}

func (c *Context) merge(old *Context) {
	if old == nil || old.Namespace == nil {
		return
	}
	if c.Namespace == nil {
		c.Namespace = NewNamespace()
	}
	c.Namespace.merge(old.Namespace)
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

	if a := os.Getenv(envPFAddress); a != "" {
		c.PortForwardAddress = a
	}
	if c.PortForwardAddress == "" {
		c.PortForwardAddress = defaultPFAddress()
	}
	if cl, err := ks.CurrentClusterName(); err == nil {
		c.ClusterName = cl
	}
	if b := os.Getenv(envFGNodeShell); b != "" {
		c.FeatureGates.NodeShell = defaultFGNodeShell()
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
