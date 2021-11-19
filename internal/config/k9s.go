package config

import (
	"github.com/derailed/k9s/internal/client"
	"os"
)

const (
	defaultRefreshRate  = 2
	defaultMaxConnRetry = 5
)

// K9s tracks K9s configuration options.
type K9s struct {
	RefreshRate       int                 `yaml:"refreshRate"`
	MaxConnRetry      int                 `yaml:"maxConnRetry"`
	EnableMouse       bool                `yaml:"enableMouse"`
	Headless          bool                `yaml:"headless"`
	Logoless          bool                `yaml:"logoless"`
	Crumbsless        bool                `yaml:"crumbsless"`
	ReadOnly          bool                `yaml:"readOnly"`
	NoIcons           bool                `yaml:"noIcons"`
	Logger            *Logger             `yaml:"logger"`
	CurrentContext    string              `yaml:"currentContext"`
	CurrentCluster    string              `yaml:"currentCluster"`
	Clusters          map[string]*Cluster `yaml:"clusters,omitempty"`
	Thresholds        Threshold           `yaml:"thresholds"`
	DumpDirPath       string              `yaml:"dumpDirPath"`
	manualRefreshRate int
	manualHeadless    *bool
	manualLogoless    *bool
	manualCrumbsless  *bool
	manualReadOnly    *bool
	manualCommand     *string
	manualDumpDirPath *string
}

// NewK9s create a new K9s configuration.
func NewK9s() *K9s {
	return &K9s{
		RefreshRate:  defaultRefreshRate,
		MaxConnRetry: defaultMaxConnRetry,
		Logger:       NewLogger(),
		Clusters:     make(map[string]*Cluster),
		Thresholds:   NewThreshold(),
		DumpDirPath:  os.TempDir(),
	}
}

// ActivateCluster initializes the active cluster is not present.
func (k *K9s) ActivateCluster() {
	if _, ok := k.Clusters[k.CurrentCluster]; ok {
		return
	}
	k.Clusters[k.CurrentCluster] = NewCluster()
}

// OverrideRefreshRate set the refresh rate manually.
func (k *K9s) OverrideRefreshRate(r int) {
	k.manualRefreshRate = r
}

// OverrideHeadless set the headlessness manually.
func (k *K9s) OverrideHeadless(b bool) {
	k.manualHeadless = &b
}

// OverrideLogoless set the logolessness manually.
func (k *K9s) OverrideLogoless(b bool) {
	k.manualLogoless = &b
}

// OverrideCrumbsless set the crumbslessness manually.
func (k *K9s) OverrideCrumbsless(b bool) {
	k.manualCrumbsless = &b
}

// OverrideReadOnly set the readonly mode manually.
func (k *K9s) OverrideReadOnly(b bool) {
	if b {
		k.manualReadOnly = &b
	}
}

// OverrideWrite set the write mode manually.
func (k *K9s) OverrideWrite(b bool) {
	if b {
		var flag bool
		k.manualReadOnly = &flag
	}
}

// OverrideCommand set the command manually.
func (k *K9s) OverrideCommand(cmd string) {
	k.manualCommand = &cmd
}

// OverrideDumpDirPath set the dump dir path manually.
func (k *K9s) OverrideDumpDirPath(path string) {
	k.manualDumpDirPath = &path
}

// IsHeadless returns headless setting.
func (k *K9s) IsHeadless() bool {
	h := k.Headless
	if k.manualHeadless != nil && *k.manualHeadless {
		h = *k.manualHeadless
	}

	return h
}

// IsLogoless returns logoless setting.
func (k *K9s) IsLogoless() bool {
	h := k.Logoless
	if k.manualLogoless != nil && *k.manualLogoless {
		h = *k.manualLogoless
	}

	return h
}

// IsCrumbsless returns crumbsless setting.
func (k *K9s) IsCrumbsless() bool {
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

// IsReadOnly returns the readonly setting.
func (k *K9s) IsReadOnly() bool {
	readOnly := k.ReadOnly
	if k.manualReadOnly != nil {
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

func (k *K9s) GetDumpDirPath() string {
	dumpDirPath := k.DumpDirPath

	if k.manualDumpDirPath != nil && *k.manualDumpDirPath != "" {
		dumpDirPath = *k.manualDumpDirPath
	}

	return dumpDirPath
}

func (k *K9s) validateDefaults() {
	if k.RefreshRate <= 0 {
		k.RefreshRate = defaultRefreshRate
	}
	if k.MaxConnRetry <= 0 {
		k.MaxConnRetry = defaultMaxConnRetry
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
