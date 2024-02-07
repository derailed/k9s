// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/derailed/k9s/internal/config/json"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
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
func (d *Dir) Load(n string, ct *api.Context) (*Config, error) {
	if ct == nil {
		return nil, errors.New("api.Context must not be nil")
	}
	var (
		path = filepath.Join(d.root, SanitizeContextSubpath(ct.Cluster, n), MainConfigFile)
		cfg  *Config
		err  error
	)
	if f, e := os.Stat(path); os.IsNotExist(e) || f.Size() == 0 {
		log.Debug().Msgf("Context config not found! Generating... %q", path)
		cfg, err = d.genConfig(path, ct)
	} else {
		log.Debug().Msgf("Found existing context config: %q", path)
		cfg, err = d.loadConfig(path)
	}

	return cfg, err
}

func (d *Dir) genConfig(path string, ct *api.Context) (*Config, error) {
	cfg := NewConfig(ct)
	if err := d.Save(path, cfg); err != nil {
		return nil, err
	}

	return cfg, nil
}

func (d *Dir) Save(path string, c *Config) error {
	d.mx.Lock()
	defer d.mx.Unlock()

	if err := EnsureDirPath(path, DefaultDirMod); err != nil {
		return err
	}
	cfg, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, cfg, DefaultFileMod)
}

func (d *Dir) loadConfig(path string) (*Config, error) {
	d.mx.Lock()
	defer d.mx.Unlock()

	bb, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err := JSONValidator.Validate(json.ContextSchema, bb); err != nil {
		return nil, fmt.Errorf("validation failed for %q: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(bb, &cfg); err != nil {
		return nil, fmt.Errorf("context-config yaml load failed: %w\n%s", err, string(bb))
	}

	return &cfg, nil
}
