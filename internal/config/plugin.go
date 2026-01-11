// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/karrick/godirwalk"
	"gopkg.in/yaml.v3"
)

type plugins map[string]Plugin

// Plugins represents a collection of plugins.
type Plugins struct {
	Plugins plugins `yaml:"plugins"`
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
func (p Plugins) Load(path string, loadExtra bool) error {
	var errs error

	// Load from global config file
	if err := p.load(AppPluginsFile); err != nil {
		errs = errors.Join(errs, err)
	}

	// Load from cluster/context config
	if err := p.load(path); err != nil {
		errs = errors.Join(errs, err)
	}

	if !loadExtra {
		return errs
	}
	// Load from XDG dirs
	const k9sPluginsDir = "k9s/plugins"
	for _, dir := range append(xdg.DataDirs, xdg.DataHome, xdg.ConfigHome) {
		path := filepath.Join(dir, k9sPluginsDir)
		if err := p.loadDir(path); err != nil {
			errs = errors.Join(errs, err)
		}
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
	scheme, err := data.JSONValidator.ValidatePlugins(bb)
	if err != nil {
		slog.Warn("Plugin schema validation failed",
			slogs.Path, path,
			slogs.Error, err,
		)
		return fmt.Errorf("plugin validation failed for %s: %w", path, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(bb))
	d.KnownFields(true)

	switch scheme {
	case json.PluginSchema:
		var o Plugin
		if err := yaml.Unmarshal(bb, &o); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		p.Plugins[strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))] = o
	case json.PluginsSchema:
		var oo Plugins
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		for k := range oo.Plugins {
			p.Plugins[k] = oo.Plugins[k]
		}
	case json.PluginMultiSchema:
		var oo plugins
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		for k := range oo {
			p.Plugins[k] = oo[k]
		}
	}

	return nil
}

func (p Plugins) loadDir(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	var errs error
	errs = errors.Join(errs, godirwalk.Walk(dir, &godirwalk.Options{
		FollowSymbolicLinks: true,
		Callback: func(path string, de *godirwalk.Dirent) error {
			if de.IsDir() || !isYamlFile(de.Name()) {
				return nil
			}
			errs = errors.Join(errs, p.load(path))
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			slog.Warn("Error at %s: %v - skipping node", slogs.Path, osPathname, slogs.Error, err)
			return godirwalk.SkipNode
		},
	}))

	return errs
}
