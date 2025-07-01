// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v3"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Dir tracks context configurations.
type Dir struct {
	root string
	mx   sync.Mutex
}

// NewDir returns a new instance.
func NewDir(root string) *Dir {
	return &Dir{
		root: root,
	}
}

// Load loads context configuration.
func (d *Dir) Load(contextName string, ct *api.Context) (*Config, error) {
	if ct == nil {
		return nil, errors.New("api.Context must not be nil")
	}

	path := filepath.Join(d.root, SanitizeContextSubpath(ct.Cluster, contextName), MainConfigFile)
	slog.Debug("[CONFIG] Loading context config from disk", slogs.Path, path, slogs.Cluster, ct.Cluster, slogs.Context, contextName)
	f, err := os.Stat(path)
	if errors.Is(err, fs.ErrPermission) {
		return nil, err
	}
	if errors.Is(err, fs.ErrNotExist) || (f != nil && f.Size() == 0) {
		slog.Debug("Context config not found! Generating..", slogs.Path, path)
		return d.genConfig(path, ct)
	}
	if err != nil {
		return nil, err
	}

	return d.loadConfig(path)
}

func (d *Dir) genConfig(path string, ct *api.Context) (*Config, error) {
	cfg := NewConfig(ct)
	if err := d.Save(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d *Dir) Save(path string, c *Config) error {
	if cfg, err := d.loadConfig(path); err == nil {
		c.Merge(cfg)
	}

	d.mx.Lock()
	defer d.mx.Unlock()

	if err := EnsureDirPath(path, DefaultDirMod); err != nil {
		return err
	}

	return SaveYAML(path, c)
}

func (d *Dir) loadConfig(path string) (*Config, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	bb, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := JSONValidator.Validate(json.ContextSchema, bb); err != nil {
		slog.Warn("Validation failed. Please update your config and restart!",
			slogs.Path, path,
			slogs.Error, err,
		)
	}

	var cfg Config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return nil, fmt.Errorf("context-config yaml load failed: %w\n%s", err, string(bb))
	}

	return &cfg, nil
}
