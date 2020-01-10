package config

import "github.com/derailed/k9s/internal/client"

const (
	defaultRefreshRate    = 2
	defaultLogRequestSize = 200
	defaultLogBufferSize  = 1000
)

// K9s tracks K9s configuration options.
type K9s struct {
	RefreshRate       int                 `yaml:"refreshRate"`
	Headless          bool                `yaml:"headless"`
	LogBufferSize     int                 `yaml:"logBufferSize"`
	LogRequestSize    int                 `yaml:"logRequestSize"`
	CurrentContext    string              `yaml:"currentContext"`
	CurrentCluster    string              `yaml:"currentCluster"`
	FullScreenLogs    bool                `yaml:"fullScreenLogs"`
	Clusters          map[string]*Cluster `yaml:"clusters,omitempty"`
	manualRefreshRate int
	manualHeadless    *bool
	manualCommand     *string
}

// NewK9s create a new K9s configuration.
func NewK9s() *K9s {
	return &K9s{
		RefreshRate:    defaultRefreshRate,
		LogBufferSize:  defaultLogBufferSize,
		LogRequestSize: defaultLogRequestSize,
		Clusters:       make(map[string]*Cluster),
	}
}

// OverrideRefreshRate set the refresh rate manually.
func (k *K9s) OverrideRefreshRate(r int) {
	k.manualRefreshRate = r
}

// OverrideHeadless set the headlessness manually.
func (k *K9s) OverrideHeadless(b bool) {
	k.manualHeadless = &b
}

// OverrideCommand set the command manually.
func (k *K9s) OverrideCommand(cmd string) {
	k.manualCommand = &cmd
}

// GetHeadless returns headless setting.
func (k *K9s) GetHeadless() bool {
	h := k.Headless
	if k.manualHeadless != nil && *k.manualHeadless {
		h = *k.manualHeadless
	}

	return h
}

// GetRefreshRate returns the current refresh rate.
func (k *K9s) GetRefreshRate() int {
	rate := k.RefreshRate
	if k.manualRefreshRate != 0 {
		rate = k.manualRefreshRate
	}

	return rate
}

// ActiveCluster returns the currently active cluster.
func (k *K9s) ActiveCluster() *Cluster {
	if k.Clusters == nil {
		k.Clusters = map[string]*Cluster{}
	}

	if c, ok := k.Clusters[k.CurrentCluster]; ok {
		return c
	}
	k.Clusters[k.CurrentCluster] = NewCluster()

	return k.Clusters[k.CurrentCluster]
}

func (k *K9s) validateDefaults() {
	if k.RefreshRate <= 0 {
		k.RefreshRate = defaultRefreshRate
	}

	if k.LogBufferSize <= 0 {
		k.LogBufferSize = defaultLogBufferSize
	}

	if k.LogRequestSize <= 0 {
		k.LogRequestSize = defaultLogRequestSize
	}
}

func (k *K9s) checkClusters(ks KubeSettings) {
	cc, err := ks.ClusterNames()
	if err != nil {
		return
	}
	for key := range k.Clusters {
		if InList(cc, key) {
			continue
		}
		if k.CurrentCluster == key {
			k.CurrentCluster = ""
		}
		delete(k.Clusters, key)
	}
}

// Validate the current configuration.
func (k *K9s) Validate(c client.Connection, ks KubeSettings) {
	k.validateDefaults()

	if k.Clusters == nil {
		k.Clusters = map[string]*Cluster{}
	}
	k.checkClusters(ks)

	if ctx, err := ks.CurrentContextName(); err == nil && len(k.CurrentContext) == 0 {
		k.CurrentContext = ctx
		k.CurrentCluster = ""
	}

	if cl, err := ks.CurrentClusterName(); err == nil && len(k.CurrentCluster) == 0 {
		k.CurrentCluster = cl
	}

	if _, ok := k.Clusters[k.CurrentCluster]; !ok {
		k.Clusters[k.CurrentCluster] = NewCluster()
	}
	k.Clusters[k.CurrentCluster].Validate(c, ks)
}
