// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v2"
)

const k9sPluginsDir = "k9s/plugins"

// Plugins represents a collection of plugins.
type Plugins struct {
	Plugins map[string]Plugin `yaml:"plugins"`
}

// Plugin describes a K9s plugin.
type Plugin struct {
	Scopes      []string `yaml:"scopes"`
	Args        []string `yaml:"args"`
	ShortCut    string   `yaml:"shortCut"`
	Pipes       []string `yaml:"pipes"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Confirm     bool     `yaml:"confirm"`
	Background  bool     `yaml:"background"`
}

func (p Plugin) String() string {
	return fmt.Sprintf("[%s] %s(%s)", p.ShortCut, p.Command, strings.Join(p.Args, " "))
}

// NewPlugins returns a new plugin.
func NewPlugins() Plugins {
	return Plugins{
		Plugins: make(map[string]Plugin),
	}
}

// Load K9s plugins.
func (p Plugins) Load(path string) error {
	var errs error

	if err := p.load(AppPluginsFile); err != nil {
		errs = errors.Join(errs, err)
	}
	if err := p.load(path); err != nil {
		errs = errors.Join(errs, err)
	}

	for _, dataDir := range xdg.DataDirs {
		if err := p.loadPluginDir(filepath.Join(dataDir, k9sPluginsDir)); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (p Plugins) loadPluginDir(dir string) error {
	pluginFiles, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	var errs error
	for _, file := range pluginFiles {
		if file.IsDir() || !isYamlFile(file.Name()) {
			continue
		}
		bb, err := os.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			errs = errors.Join(errs, err)
		}
		var plugin Plugin
		if err = yaml.Unmarshal(bb, &plugin); err != nil {
			return err
		}
		p.Plugins[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = plugin
	}

	return errs
}

func (p *Plugins) load(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.PluginsSchema, bb); err != nil {
		return fmt.Errorf("validation failed for %q: %w", path, err)
	}
	var pp Plugins
	if err := yaml.Unmarshal(bb, &pp); err != nil {
		return err
	}
	for k, v := range pp.Plugins {
		p.Plugins[k] = v
	}

	return nil
}
