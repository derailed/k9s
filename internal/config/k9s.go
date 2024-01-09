// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"
	"path/filepath"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
)

// K9s tracks K9s configuration options.
type K9s struct {
	LiveViewAutoRefresh bool       `json:"liveViewAutoRefresh" yaml:"liveViewAutoRefresh"`
	ScreenDumpDir       string     `json:"screenDumpDir" yaml:"screenDumpDir,omitempty"`
	RefreshRate         int        `json:"refreshRate" yaml:"refreshRate"`
	MaxConnRetry        int        `json:"maxConnRetry" yaml:"maxConnRetry"`
	ReadOnly            bool       `json:"readOnly" yaml:"readOnly"`
	NoExitOnCtrlC       bool       `json:"noExitOnCtrlC" yaml:"noExitOnCtrlC"`
	UI                  UI         `json:"ui" yaml:"ui"`
	SkipLatestRevCheck  bool       `json:"skipLatestRevCheck" yaml:"skipLatestRevCheck"`
	DisablePodCounting  bool       `json:"disablePodCounting" yaml:"disablePodCounting"`
	ShellPod            ShellPod   `json:"shellPod" yaml:"shellPod"`
	ImageScans          ImageScans `json:"imageScans" yaml:"imageScans"`
	Logger              Logger     `json:"logger" yaml:"logger"`
	Thresholds          Threshold  `json:"thresholds" yaml:"thresholds"`
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
	mx                  sync.RWMutex
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
	k.mx.Lock()
	defer k.mx.Unlock()

	k.conn = conn
}

// Save saves the k9s config to dis.
func (k *K9s) Save() error {
	if k.activeConfig == nil {
		return fmt.Errorf("save failed. no active config detected")
	}
	path := filepath.Join(
		AppContextsDir,
		data.SanitizeContextSubpath(k.activeConfig.Context.ClusterName, k.activeContextName),
		data.MainConfigFile,
	)

	return k.activeConfig.Save(path)
}

// Merge merges k9s configs.
func (k *K9s) Merge(k1 *K9s) {
	if k1 == nil {
		return
	}

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
	k.Logger = k1.Logger
	k.ImageScans = k1.ImageScans
	k.Thresholds = k1.Thresholds
}

// AppScreenDumpDir fetch screen dumps dir.
func (k *K9s) AppScreenDumpDir() string {
	d := k.ScreenDumpDir
	if isStringSet(k.manualScreenDumpDir) {
		d = *k.manualScreenDumpDir
		k.ScreenDumpDir = d
	}
	if d == "" {
		d = AppDumpsDir
	}

	return d
}

// ContextScreenDumpDir fetch context specific screen dumps dir.
func (k *K9s) ContextScreenDumpDir() string {
	return filepath.Join(k.AppScreenDumpDir(), k.contextPath())
}

func (k *K9s) contextPath() string {
	if k.activeConfig == nil {
		return "na"
	}

	return data.SanitizeContextSubpath(
		k.activeConfig.Context.ClusterName,
		k.ActiveContextName(),
	)
}

// Reset resets configuration and context.
func (k *K9s) Reset() {
	k.activeConfig, k.activeContextName = nil, ""
}

// ActiveContextNamespace fetch the context active ns.
func (k *K9s) ActiveContextNamespace() (string, error) {
	act, err := k.ActiveContext()
	if err != nil {
		return "", err
	}

	return act.Namespace.Active, nil
}

// ActiveContextName returns the active context name.
func (k *K9s) ActiveContextName() string {
	k.mx.RLock()
	defer k.mx.RUnlock()

	return k.activeContextName
}

// ActiveContext returns the currently active context.
func (k *K9s) ActiveContext() (*data.Context, error) {
	var ac *data.Config
	k.mx.RLock()
	ac = k.activeConfig
	k.mx.RUnlock()

	if ac != nil && ac.Context != nil {
		return ac.Context, nil
	}
	ct, err := k.ActivateContext(k.activeContextName)
	if err != nil {
		return nil, err
	}

	return ct, nil
}

// ActivateContext initializes the active context if not present.
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

	k.Validate(k.conn, k.ks)
	// If the context specifies a namespace, use it!
	if ns := ct.Namespace; ns != client.BlankNamespace {
		k.activeConfig.Context.Namespace.Active = ns
	} else {
		k.activeConfig.Context.Namespace.Active = client.DefaultNamespace
	}
	if k.activeConfig.Context == nil {
		return nil, fmt.Errorf("context activation failed for: %s", n)
	}

	return k.activeConfig.Context, nil
}

// Reload reloads the context config from disk.
func (k *K9s) Reload() error {
	k.mx.Lock()
	defer k.mx.Unlock()

	ct, err := k.ks.GetContext(k.activeContextName)
	if err != nil {
		return err
	}
	k.activeConfig, err = k.dir.Load(k.activeContextName, ct)
	if err != nil {
		return err
	}
	k.activeConfig.Validate(k.conn, k.ks)

	return nil
}

// Override overrides k9s config from cli args.
func (k *K9s) Override(k9sFlags *Flags) {
	if k9sFlags.RefreshRate != nil && *k9sFlags.RefreshRate != DefaultRefreshRate {
		k.manualRefreshRate = *k9sFlags.RefreshRate
	}

	k.manualHeadless = k9sFlags.Headless
	k.manualLogoless = k9sFlags.Logoless
	k.manualCrumbsless = k9sFlags.Crumbsless
	if k9sFlags.ReadOnly != nil && *k9sFlags.ReadOnly {
		k.manualReadOnly = k9sFlags.ReadOnly
	}
	if k9sFlags.Write != nil && *k9sFlags.Write {
		var false bool
		k.manualReadOnly = &false
	}
	k.manualCommand = k9sFlags.Command
	k.manualScreenDumpDir = k9sFlags.ScreenDumpDir
}

// IsHeadless returns headless setting.
func (k *K9s) IsHeadless() bool {
	if isBoolSet(k.manualHeadless) {
		return true
	}

	return k.UI.Headless
}

// IsLogoless returns logoless setting.
func (k *K9s) IsLogoless() bool {
	if isBoolSet(k.manualLogoless) {
		return true
	}

	return k.UI.Logoless
}

// IsCrumbsless returns crumbsless setting.
func (k *K9s) IsCrumbsless() bool {
	if isBoolSet(k.manualCrumbsless) {
		return true
	}

	return k.UI.Crumbsless
}

// GetRefreshRate returns the current refresh rate.
func (k *K9s) GetRefreshRate() int {
	if k.manualRefreshRate != 0 {
		return k.manualRefreshRate
	}

	return k.RefreshRate
}

// IsReadOnly returns the readonly setting.
func (k *K9s) IsReadOnly() bool {
	k.mx.RLock()
	defer k.mx.RUnlock()

	ro := k.ReadOnly
	if k.activeConfig != nil && k.activeConfig.Context.ReadOnly != nil {
		ro = *k.activeConfig.Context.ReadOnly
	}
	if k.manualReadOnly != nil {
		ro = true
	}

	return ro
}

// Validate the current configuration.
func (k *K9s) Validate(c client.Connection, ks data.KubeSettings) {
	if k.RefreshRate <= 0 {
		k.RefreshRate = defaultRefreshRate
	}
	if k.MaxConnRetry <= 0 {
		k.MaxConnRetry = defaultMaxConnRetry
	}

	if k.activeConfig == nil {
		if n, err := ks.CurrentContextName(); err == nil {
			_, _ = k.ActivateContext(n)
		}
	}
	k.ShellPod = k.ShellPod.Validate()
	k.Logger = k.Logger.Validate()
	k.Thresholds = k.Thresholds.Validate()

	if k.activeConfig != nil {
		k.activeConfig.Validate(c, ks)
	}
}
