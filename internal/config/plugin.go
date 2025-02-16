// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"io/fs"
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
	Scopes          []string `yaml:"scopes"`
	Args            []string `yaml:"args"`
	ShortCut        string   `yaml:"shortCut"`
	Override        bool     `yaml:"override"`
	Pipes           []string `yaml:"pipes"`
	Description     string   `yaml:"description"`
	Command         string   `yaml:"command"`
	Confirm         bool     `yaml:"confirm"`
	Background      bool     `yaml:"background"`
	Dangerous       bool     `yaml:"dangerous"`
	OverwriteOutput bool     `yaml:"overwriteOutput"`
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

	for _, dataDir := range append(xdg.DataDirs, xdg.DataHome, xdg.ConfigHome) {
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
		fileName := filepath.Join(dir, file.Name())
		fileContent, err := os.ReadFile(fileName)
		if err != nil {
			errs = errors.Join(errs, err)
		}
		var plugin Plugin
		if err = yaml.UnmarshalStrict(fileContent, &plugin); err != nil {
			var plugins Plugins
			if err = yaml.UnmarshalStrict(fileContent, &plugins); err != nil {
				return fmt.Errorf("cannot parse %s into either a single plugin nor plugins: %w", fileName, err)
			}
			for name, plugin := range plugins.Plugins {
				p.Plugins[name] = plugin
			}
			continue
		}
		p.Plugins[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = plugin
	}

	return errs
}

func (p *Plugins) load(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
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
