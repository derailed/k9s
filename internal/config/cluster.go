package config

import "github.com/derailed/k9s/internal/client"

// DefaultPFAddress specifies the default PortForward host address.
const DefaultPFAddress = "localhost"

// Cluster tracks K9s cluster configuration.
type Cluster struct {
	Namespace          *Namespace    `yaml:"namespace"`
	View               *View         `yaml:"view"`
	FeatureGates       *FeatureGates `yaml:"featureGates"`
	ShellPod           *ShellPod     `yaml:"shellPod"`
	PortForwardAddress string        `yaml:"portForwardAddress"`
}

// NewCluster creates a new cluster configuration.
func NewCluster() *Cluster {
	return &Cluster{
		Namespace:          NewNamespace(),
		View:               NewView(),
		PortForwardAddress: DefaultPFAddress,
		FeatureGates:       NewFeatureGates(),
		ShellPod:           NewShellPod(),
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
	if c.Namespace.Active == client.AllNamespaces {
		c.Namespace.Active = client.NamespaceAll
	}

	if c.FeatureGates == nil {
		c.FeatureGates = NewFeatureGates()
	}

	if c.View == nil {
		c.View = NewView()
	}
	c.View.Validate()

	if c.ShellPod == nil {
		c.ShellPod = NewShellPod()
	}
	c.ShellPod.Validate(conn, ks)
}
