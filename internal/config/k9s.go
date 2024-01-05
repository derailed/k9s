// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"path/filepath"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
)

// K9s tracks K9s configuration options.
type K9s struct {
	LiveViewAutoRefresh bool        `yaml:"liveViewAutoRefresh"`
	ScreenDumpDir       string      `yaml:"screenDumpDir,omitempty"`
	RefreshRate         int         `yaml:"refreshRate"`
	MaxConnRetry        int         `yaml:"maxConnRetry"`
	ReadOnly            bool        `yaml:"readOnly"`
	NoExitOnCtrlC       bool        `yaml:"noExitOnCtrlC"`
	UI                  UI          `yaml:"ui"`
	SkipLatestRevCheck  bool        `yaml:"skipLatestRevCheck"`
	DisablePodCounting  bool        `yaml:"disablePodCounting"`
	ShellPod            *ShellPod   `yaml:"shellPod"`
	ImageScans          *ImageScans `yaml:"imageScans"`
	Logger              *Logger     `yaml:"logger"`
	Thresholds          Threshold   `yaml:"thresholds"`
	manualRefreshRate   int
	manualHeadless      *bool
	manualLogoless      *bool
	manualCrumbsless    *bool
	manualReadOnly      *bool
	manualCommand       *string
	manualScreenDumpDir *string
	dir                 *data.Dir
	activeContextName   string
	activeConfig        *data.Config
	conn                client.Connection
	ks                  data.KubeSettings
}

// NewK9s create a new K9s configuration.
func NewK9s(conn client.Connection, ks data.KubeSettings) *K9s {
	return &K9s{
		RefreshRate:   defaultRefreshRate,
		MaxConnRetry:  defaultMaxConnRetry,
		ScreenDumpDir: AppDumpsDir,
		Logger:        NewLogger(),
		Thresholds:    NewThreshold(),
		ShellPod:      NewShellPod(),
		ImageScans:    NewImageScans(),
		dir:           data.NewDir(AppContextsDir),
		conn:          conn,
		ks:            ks,
	}
}

func (k *K9s) resetConnection(conn client.Connection) {
	k.conn = conn
}

// Save saves the k9s config to dis.
func (k *K9s) Save() error {
	if k.activeConfig != nil {
		path := filepath.Join(
			AppContextsDir,
			data.SanitizeContextSubpath(k.activeConfig.Context.ClusterName, k.activeContextName),
			data.MainConfigFile,
		)
		return k.activeConfig.Save(path)
	}

	return nil
}

// Refine merges k9s configs.
func (k *K9s) Refine(k1 *K9s) {
	k.LiveViewAutoRefresh = k1.LiveViewAutoRefresh
	k.ScreenDumpDir = k1.ScreenDumpDir
	k.RefreshRate = k1.RefreshRate
	k.MaxConnRetry = k1.MaxConnRetry
	k.ReadOnly = k1.ReadOnly
	k.NoExitOnCtrlC = k1.NoExitOnCtrlC
	k.UI = k1.UI
	k.SkipLatestRevCheck = k1.SkipLatestRevCheck
	k.DisablePodCounting = k1.DisablePodCounting
	k.ShellPod = k1.ShellPod
	k.ImageScans = k1.ImageScans
	k.Logger = k1.Logger
	k.Thresholds = k1.Thresholds
}

// Override overrides k9s config from cli args.
func (k *K9s) Override(k9sFlags *Flags) {
	if *k9sFlags.RefreshRate != DefaultRefreshRate {
		k.OverrideRefreshRate(*k9sFlags.RefreshRate)
	}

	k.OverrideHeadless(*k9sFlags.Headless)
	k.OverrideLogoless(*k9sFlags.Logoless)
	k.OverrideCrumbsless(*k9sFlags.Crumbsless)
	k.OverrideReadOnly(*k9sFlags.ReadOnly)
	k.OverrideWrite(*k9sFlags.Write)
	k.OverrideCommand(*k9sFlags.Command)
	k.OverrideScreenDumpDir(*k9sFlags.ScreenDumpDir)
}

// OverrideScreenDumpDir set the screen dump dir manually.
func (k *K9s) OverrideScreenDumpDir(dir string) {
	k.manualScreenDumpDir = &dir
}

// GetScreenDumpDir fetch screen dumps dir.
func (k *K9s) GetScreenDumpDir() string {
	screenDumpDir := k.ScreenDumpDir
	if k.manualScreenDumpDir != nil && *k.manualScreenDumpDir != "" {
		screenDumpDir = *k.manualScreenDumpDir
	}
	if screenDumpDir == "" {
		screenDumpDir = AppDumpsDir
	}

	return screenDumpDir
}

// Reset resets configuration and context.
func (k *K9s) Reset() {
	k.activeConfig, k.activeContextName = nil, ""
}

// ActiveScreenDumpsDir fetch context specific screen dumps dir.
func (k *K9s) ActiveScreenDumpsDir() string {
	return filepath.Join(k.GetScreenDumpDir(), k.ActiveContextDir())
}

// ActiveContextDir fetch current cluster/context path.
func (k *K9s) ActiveContextDir() string {
	if k.activeConfig == nil {
		return "na"
	}

	return data.SanitizeContextSubpath(
		k.activeConfig.Context.ClusterName,
		k.ActiveContextName(),
	)
}

// ActiveContextNamespace fetch the context active ns.
func (k *K9s) ActiveContextNamespace() (string, error) {
	if k.activeConfig != nil {
		return k.activeConfig.Context.Namespace.Active, nil
	}

	return "", errors.New("context config is not set")
}

// ActiveContextName returns the active context name.
func (k *K9s) ActiveContextName() string {
	return k.activeContextName
}

// ActiveContext returns the currently active context.
func (k *K9s) ActiveContext() (*data.Context, error) {
	if k.activeConfig != nil {
		if k.activeConfig.Context == nil {
			ct, err := k.ks.CurrentContext()
			if err != nil {
				return nil, err
			}
			k.activeConfig.Context = data.NewContextFromConfig(ct)
		}
		return k.activeConfig.Context, nil
	}

	ct, err := k.ActivateContext(k.activeContextName)
	if err != nil {
		return nil, err
	}

	return ct, nil
}

// ActivateContext initializes the active context is not present.
func (k *K9s) ActivateContext(n string) (*data.Context, error) {
	k.activeContextName = n
	ct, err := k.ks.GetContext(n)
	if err != nil {
		return nil, err
	}
	k.activeConfig, err = k.dir.Load(n, ct)
	if err != nil {
		return nil, err
	}
	// If the context specifies a default namespace, use it!
	if k.conn != nil {
		k.Validate(k.conn, k.ks)
		if ns := k.conn.ActiveNamespace(); ns != client.BlankNamespace {
			k.activeConfig.Context.Namespace.Active = ns
		} else {
			k.activeConfig.Context.Namespace.Active = client.DefaultNamespace
		}
	}

	return k.activeConfig.Context, nil
}

// Reload reloads the active config from disk.
func (k *K9s) Reload() error {
	ct, err := k.ks.GetContext(k.activeContextName)
	if err != nil {
		return err
	}

	k.activeConfig, err = k.dir.Load(k.activeContextName, ct)
	if err != nil {
		return err
	}

	return nil
}

// OverrideRefreshRate set the refresh rate manually.
func (k *K9s) OverrideRefreshRate(r int) {
	k.manualRefreshRate = r
}

// OverrideHeadless toggle the header manually.
func (k *K9s) OverrideHeadless(b bool) {
	k.manualHeadless = &b
}

// OverrideLogoless toggle the k9s logo manually.
func (k *K9s) OverrideLogoless(b bool) {
	k.manualLogoless = &b
}

// OverrideCrumbsless tooh the crumbslessness manually.
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

// IsHeadless returns headless setting.
func (k *K9s) IsHeadless() bool {
	h := k.UI.Headless
	if k.manualHeadless != nil && *k.manualHeadless {
		h = *k.manualHeadless
	}

	return h
}

// IsLogoless returns logoless setting.
func (k *K9s) IsLogoless() bool {
	h := k.UI.Logoless
	if k.manualLogoless != nil && *k.manualLogoless {
		h = *k.manualLogoless
	}

	return h
}

// IsCrumbsless returns crumbsless setting.
func (k *K9s) IsCrumbsless() bool {
	h := k.UI.Crumbsless
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
	if k.activeConfig != nil && k.activeConfig.Context.ReadOnly {
		readOnly = true
	}

	return readOnly
}

func (k *K9s) validateDefaults() {
	if k.RefreshRate <= 0 {
		k.RefreshRate = defaultRefreshRate
	}
	if k.MaxConnRetry <= 0 {
		k.MaxConnRetry = defaultMaxConnRetry
	}
}

// Validate the current configuration.
func (k *K9s) Validate(c client.Connection, ks data.KubeSettings) {
	k.validateDefaults()
	if k.activeConfig == nil {
		if n, err := ks.CurrentContextName(); err == nil {
			_, _ = k.ActivateContext(n)
		}
	}
	if k.ImageScans == nil {
		k.ImageScans = NewImageScans()
	}
	if k.ShellPod == nil {
		k.ShellPod = NewShellPod()
	}
	k.ShellPod.Validate()

	if k.Logger == nil {
		k.Logger = NewLogger()
	} else {
		k.Logger.Validate()
	}
	if k.Thresholds == nil {
		k.Thresholds = NewThreshold()
	}
	k.Thresholds.Validate()

	if k.activeConfig != nil {
		k.activeConfig.Validate(c, ks)
	}
}
