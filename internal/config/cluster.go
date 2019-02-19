package config

// Cluster tracks K9s cluster configuration.
type Cluster struct {
	Namespace *Namespace `yaml:"namespace"`
	View      *View      `yaml:"view"`
}

// NewCluster creates a new cluster configuration.
func NewCluster() *Cluster {
	return &Cluster{Namespace: NewNamespace(), View: NewView()}
}

// Validate a cluster config.
func (c *Cluster) Validate(ks KubeSettings) {
	if c.Namespace == nil {
		c.Namespace = NewNamespace()
	}
	c.Namespace.Validate(ks)

	if c.View == nil {
		c.View = NewView()
	}
	c.View.Validate()
}
