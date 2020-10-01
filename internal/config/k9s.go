package config

import "github.com/derailed/k9s/internal/client"

const defaultRefreshRate = 2

// K9s tracks K9s configuration options.
type K9s struct {
	RefreshRate       int                 `yaml:"refreshRate"`
	EnableMouse       bool                `yaml:"enableMouse"`
	Headless          bool                `yaml:"headless"`
	Crumbsless        bool                `yaml:"crumbsless"`
	ReadOnly          bool                `yaml:"readOnly"`
	NoIcons           bool                `yaml:"noIcons"`
	Logger            *Logger             `yaml:"logger"`
	CurrentContext    string              `yaml:"currentContext"`
	CurrentCluster    string              `yaml:"currentCluster"`
	Clusters          map[string]*Cluster `yaml:"clusters,omitempty"`
	Thresholds        Threshold           `yaml:"thresholds"`
	manualRefreshRate int
	manualHeadless    *bool
	manualCrumbsless  *bool
	manualReadOnly    *bool
	manualCommand     *string
}

// NewK9s create a new K9s configuration.
func NewK9s() *K9s {
	return &K9s{
		RefreshRate: defaultRefreshRate,
		Logger:      NewLogger(),
		Clusters:    make(map[string]*Cluster),
		Thresholds:  NewThreshold(),
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

// OverrideCrumbsless set the headlessness manually.
func (k *K9s) OverrideCrumbsless(b bool) {
	k.manualCrumbsless = &b
}

// OverrideReadOnly set the readonly mode manually.
func (k *K9s) OverrideReadOnly(b bool) {
	k.manualReadOnly = &b
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

// GetCrumbsless returns crumbsless setting.
func (k *K9s) GetCrumbsless() bool {
	h := k.Crumbsless
	if k.manualCrumbsless != nil && *k.manualCrumbsless {
		h = *k.manualCrumbsless
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

// GetReadOnly returns the readonly setting.
func (k *K9s) GetReadOnly() bool {
	readOnly := k.ReadOnly
	if k.manualReadOnly != nil && *k.manualReadOnly {
		readOnly = *k.manualReadOnly
	}
	return readOnly
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
}

func (k *K9s) validateClusters(c client.Connection, ks KubeSettings) {
	cc, err := ks.ClusterNames()
	if err != nil {
		return
	}
	for key := range k.Clusters {
		k.Clusters[key].Validate(c, ks)
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
	k.validateClusters(c, ks)

	if k.Logger == nil {
		k.Logger = NewLogger()
	} else {
		k.Logger.Validate(c, ks)
	}
	if k.Thresholds == nil {
		k.Thresholds = NewThreshold()
	}
	k.Thresholds.Validate(c, ks)

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
