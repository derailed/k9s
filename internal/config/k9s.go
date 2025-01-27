// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/rs/zerolog/log"
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

// Save saves the k9s config to disk.
func (k *K9s) Save(force bool) error {
	if k.getActiveConfig() == nil {
		log.Warn().Msgf("Save failed. no active config detected")
		return nil
	}
	path := filepath.Join(
		AppContextsDir,
		data.SanitizeContextSubpath(k.activeConfig.Context.GetClusterName(), k.getActiveContextName()),
		data.MainConfigFile,
	)
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) || force {
		return k.dir.Save(path, k.getActiveConfig())
	}

	return nil
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
	if k1.Thresholds != nil {
		k.Thresholds = k1.Thresholds
	}
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
	if k.getActiveConfig() == nil {
		return "na"
	}

	return data.SanitizeContextSubpath(
		k.getActiveConfig().Context.GetClusterName(),
		k.ActiveContextName(),
	)
}

// Reset resets configuration and context.
func (k *K9s) Reset() {
	k.setActiveConfig(nil)
	k.setActiveContextName("")
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
	return k.getActiveContextName()
}

// ActiveContext returns the currently active context.
func (k *K9s) ActiveContext() (*data.Context, error) {
	if cfg := k.getActiveConfig(); cfg != nil && cfg.Context != nil {
		return cfg.Context, nil
	}
	ct, err := k.ActivateContext(k.getActiveContextName())
	if err != nil {
		return nil, err
	}

	return ct, nil
}

func (k *K9s) setActiveConfig(c *data.Config) {
	k.mx.Lock()
	defer k.mx.Unlock()

	k.activeConfig = c
}

func (k *K9s) getActiveConfig() *data.Config {
	k.mx.RLock()
	defer k.mx.RUnlock()

	return k.activeConfig
}

func (k *K9s) setActiveContextName(n string) {
	k.mx.Lock()
	defer k.mx.Unlock()

	k.activeContextName = n
}

func (k *K9s) getActiveContextName() string {
	k.mx.RLock()
	defer k.mx.RUnlock()

	return k.activeContextName
}

// ActivateContext initializes the active context if not present.
func (k *K9s) ActivateContext(n string) (*data.Context, error) {
	k.setActiveContextName(n)
	ct, err := k.ks.GetContext(n)
	if err != nil {
		return nil, err
	}

	cfg, err := k.dir.Load(n, ct)
	if err != nil {
		return nil, err
	}
	k.setActiveConfig(cfg)

	if cfg.Context.Proxy != nil {
		k.ks.SetProxy(func(*http.Request) (*url.URL, error) {
			log.Debug().Msgf("[Proxy]: %s", cfg.Context.Proxy.Address)
			return url.Parse(cfg.Context.Proxy.Address)
		})

		if k.conn != nil && k.conn.Config() != nil {
			// We get on this branch when the user switches the context and k9s
			// already has an API connection object so we just set the proxy to
			// avoid recreation using client.InitConnection
			k.conn.Config().SetProxy(func(*http.Request) (*url.URL, error) {
				log.Debug().Msgf("[Proxy]: %s", cfg.Context.Proxy.Address)
				return url.Parse(cfg.Context.Proxy.Address)
			})

			if !k.conn.CheckConnectivity() {
				return nil, fmt.Errorf("unable to connect to context %q", n)
			}
		}
	}

	k.Validate(k.conn, k.ks)
	// If the context specifies a namespace, use it!
	if ns := ct.Namespace; ns != client.BlankNamespace {
		k.getActiveConfig().Context.Namespace.Active = ns
	} else if k.activeConfig.Context.Namespace.Active == "" {
		k.getActiveConfig().Context.Namespace.Active = client.DefaultNamespace
	}
	if k.getActiveConfig().Context == nil {
		return nil, fmt.Errorf("context activation failed for: %s", n)
	}

	return k.getActiveConfig().Context, nil
}

// Reload reloads the context config from disk.
func (k *K9s) Reload() error {
	ct, err := k.ks.GetContext(k.getActiveContextName())
	if err != nil {
		return err
	}

	cfg, err := k.dir.Load(k.getActiveContextName(), ct)
	if err != nil {
		return err
	}
	k.setActiveConfig(cfg)
	k.getActiveConfig().Validate(k.conn, k.ks)

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
	if IsBoolSet(k.manualHeadless) {
		return true
	}

	return k.UI.Headless
}

// IsLogoless returns logoless setting.
func (k *K9s) IsLogoless() bool {
	if IsBoolSet(k.manualLogoless) {
		return true
	}

	return k.UI.Logoless
}

// IsCrumbsless returns crumbsless setting.
func (k *K9s) IsCrumbsless() bool {
	if IsBoolSet(k.manualCrumbsless) {
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
	ro := k.ReadOnly
	if cfg := k.getActiveConfig(); cfg != nil && cfg.Context.ReadOnly != nil {
		ro = *cfg.Context.ReadOnly
	}
	if k.manualReadOnly != nil {
		ro = *k.manualReadOnly
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

	if k.getActiveConfig() == nil {
		if n, err := ks.CurrentContextName(); err == nil {
			_, _ = k.ActivateContext(n)
		}
	}
	k.ShellPod = k.ShellPod.Validate()
	k.Logger = k.Logger.Validate()
	k.Thresholds = k.Thresholds.Validate()

	if cfg := k.getActiveConfig(); cfg != nil {
		cfg.Validate(c, ks)
	}
}
