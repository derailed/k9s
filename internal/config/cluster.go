// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import "github.com/derailed/k9s/internal/client"

// DefaultPFAddress specifies the default PortForward host address.
const DefaultPFAddress = "localhost"

// ContextSkin tracks context specific skins when cluster spans multi contexts.
type ContextSkin struct {
	Name string `yaml:"name"`
	Skin string `yaml:"skin"`
}

// Cluster tracks K9s cluster configuration.
type Cluster struct {
	Namespace          *Namespace    `yaml:"namespace"`
	View               *View         `yaml:"view"`
	Skin               string        `yaml:"skin,omitempty"`
	ContextSkins       []ContextSkin `yaml:"skinContexts,omitempty"`
	FeatureGates       *FeatureGates `yaml:"featureGates"`
	PortForwardAddress string        `yaml:"portForwardAddress"`
}

// NewCluster creates a new cluster configuration.
func NewCluster() *Cluster {
	return &Cluster{
		Namespace:          NewNamespace(),
		View:               NewView(),
		PortForwardAddress: DefaultPFAddress,
		FeatureGates:       NewFeatureGates(),
	}
}

// Validate a cluster config.
func (c *Cluster) Validate(conn client.Connection, ks KubeSettings) {
	if c.PortForwardAddress == "" {
		c.PortForwardAddress = DefaultPFAddress
	}

	if c.Namespace == nil {
		c.Namespace = NewNamespace()
	}
	if c.Namespace.Active == client.BlankNamespace {
		c.Namespace.Active = client.NamespaceAll
	}

	if c.FeatureGates == nil {
		c.FeatureGates = NewFeatureGates()
	}

	if c.View == nil {
		c.View = NewView()
	}
	c.View.Validate()
}
