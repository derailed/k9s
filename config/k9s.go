package config

const (
	refreshRate   = 2
	logBufferSize = 200
)

// K9s tracks K9s configuration options.
type K9s struct {
	RefreshRate   int     `yaml:"refreshRate"`
	LogBufferSize int     `yaml:"logBufferSize"`
	Context      *Context `yaml:"context"`
}

// NewK9s create a new K9s configuration.
func NewK9s() *K9s {
	return &K9s{
		RefreshRate:   refreshRate,
		LogBufferSize: logBufferSize,
		Context:       NewContext(),
	}
}

// ActiveCluster return the current Cluster config.
func (k K9s) ActiveCluster() *Cluster {
	if k.Context == nil {
		k.Context = NewContext()
	}
	return k.Context.ActiveCluster()
}

// Validate the configuration
func (k K9s) Validate(ci ClusterInfo) {
	if k.RefreshRate <= 0 {
		k.RefreshRate = refreshRate
	}

	if k.LogBufferSize <= 0 {
		k.LogBufferSize = logBufferSize
	}

	if k.Context == nil {
		k.Context = NewContext()
	}
	k.Context.Validate(ci)
}
