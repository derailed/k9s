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
	"slices"
	"strconv"
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

// PluginSource identifies where a plugin was loaded from.
type PluginSource string

const (
	PluginSourceGlobal      PluginSource = "global"
	PluginSourceContext     PluginSource = "context"
	PluginSourceXDGConfig   PluginSource = "xdg-config"
	PluginSourceXDGDataHome PluginSource = "xdg-data-home"
	PluginSourceXDGDataDir  PluginSource = "xdg-data-dir"
)

// PluginEntry tracks the effective plugin plus the file it came from.
type PluginEntry struct {
	Name   string
	Path   string
	Source PluginSource
	Plugin Plugin
}

// PluginCatalog represents all effective plugins keyed by name.
type PluginCatalog struct {
	Entries map[string]PluginEntry
}

// PluginInputType represents the type of input field.
type PluginInputType string

const (
	InputTypeString   PluginInputType = "string"
	InputTypeNumber   PluginInputType = "number"
	InputTypeBool     PluginInputType = "bool"
	InputTypeDropdown PluginInputType = "dropdown"
)

// PluginInput describes an input field for a plugin.
type PluginInput struct {
	Name     string          `yaml:"name"`
	Label    string          `yaml:"label"`
	Type     PluginInputType `yaml:"type"`
	Required bool            `yaml:"required"`
	Default  string          `yaml:"default"`
	Options  []string        `yaml:"options"`
}

// Plugin describes a K9s plugin.
type Plugin struct {
	Scopes          []string      `yaml:"scopes"`
	Args            []string      `yaml:"args"`
	ShortCut        string        `yaml:"shortCut"`
	Override        bool          `yaml:"override"`
	Pipes           []string      `yaml:"pipes"`
	Description     string        `yaml:"description"`
	Command         string        `yaml:"command"`
	Confirm         *bool         `yaml:"confirm"`
	Background      bool          `yaml:"background"`
	Dangerous       bool          `yaml:"dangerous"`
	OverwriteOutput bool          `yaml:"overwriteOutput"`
	Inputs          []PluginInput `yaml:"inputs"`
}

func (p Plugin) String() string {
	return fmt.Sprintf("[%s] %s(%s)", p.ShortCut, p.Command, strings.Join(p.Args, " "))
}

// ShouldConfirm returns whether the plugin should show a confirmation dialog.
// Defaults to true when inputs are defined, false otherwise.
func (p *Plugin) ShouldConfirm() bool {
	if p.Confirm != nil {
		return *p.Confirm
	}
	return len(p.Inputs) > 0
}

// Validate checks the plugin configuration for errors.
func (p *Plugin) Validate() error {
	seen := make(map[string]struct{}, len(p.Inputs))
	for _, input := range p.Inputs {
		if _, ok := seen[input.Name]; ok {
			return fmt.Errorf("duplicate input name %q", input.Name)
		}
		seen[input.Name] = struct{}{}

		if input.Default == "" {
			continue
		}

		switch input.Type {
		case InputTypeDropdown:
			if !slices.Contains(input.Options, input.Default) {
				return fmt.Errorf("default value %q for input %q is not a valid option", input.Default, input.Name)
			}
		case InputTypeBool:
			if input.Default != "true" && input.Default != "false" {
				return fmt.Errorf("default value %q for bool input %q must be \"true\" or \"false\"", input.Default, input.Name)
			}
		case InputTypeNumber:
			if _, err := strconv.ParseFloat(input.Default, 64); err != nil {
				return fmt.Errorf("default value %q for number input %q is not a valid number", input.Default, input.Name)
			}
		}
	}

	return nil
}

// NewPlugins returns a new plugin.
func NewPlugins() Plugins {
	return Plugins{
		Plugins: make(map[string]Plugin),
	}
}

// NewPluginCatalog returns a new plugin catalog.
func NewPluginCatalog() PluginCatalog {
	return PluginCatalog{
		Entries: make(map[string]PluginEntry),
	}
}

// Load K9s plugins.
func (p Plugins) Load(path string, loadExtra bool) error {
	return loadPlugins(path, loadExtra, func(name string, plugin Plugin, _ string, _ PluginSource) {
		p.Plugins[name] = plugin
	})
}

// Load loads the effective plugin catalog including source metadata.
func (p PluginCatalog) Load(path string, loadExtra bool) error {
	return loadPlugins(path, loadExtra, func(name string, plugin Plugin, path string, source PluginSource) {
		p.Entries[name] = PluginEntry{
			Name:   name,
			Path:   path,
			Source: source,
			Plugin: plugin,
		}
	})
}

type pluginSink func(name string, plugin Plugin, path string, source PluginSource)

type pluginRoot struct {
	path   string
	source PluginSource
}

func loadPlugins(path string, loadExtra bool, sink pluginSink) error {
	var errs error

	// Load from global config file
	if err := loadPluginFile(AppPluginsFile, PluginSourceGlobal, sink); err != nil {
		errs = errors.Join(errs, err)
	}

	// Load from cluster/context config
	if err := loadPluginFile(path, PluginSourceContext, sink); err != nil {
		errs = errors.Join(errs, err)
	}

	if !loadExtra {
		return errs
	}
	for _, root := range xdgPluginRoots() {
		if err := loadPluginDir(root.path, root.source, sink); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func xdgPluginRoots() []pluginRoot {
	const k9sPluginsDir = "k9s/plugins"

	roots := make([]pluginRoot, 0, len(xdg.DataDirs)+2)
	for _, dir := range xdg.DataDirs {
		if dir == "" {
			continue
		}
		roots = append(roots, pluginRoot{
			path:   filepath.Join(dir, k9sPluginsDir),
			source: PluginSourceXDGDataDir,
		})
	}
	if xdg.DataHome != "" {
		roots = append(roots, pluginRoot{
			path:   filepath.Join(xdg.DataHome, k9sPluginsDir),
			source: PluginSourceXDGDataHome,
		})
	}
	if xdg.ConfigHome != "" {
		roots = append(roots, pluginRoot{
			path:   filepath.Join(xdg.ConfigHome, k9sPluginsDir),
			source: PluginSourceXDGConfig,
		})
	}

	return roots
}

func loadPluginFile(path string, source PluginSource, sink pluginSink) error {
	if path == "" {
		return nil
	}
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
		if err := o.Validate(); err != nil {
			return fmt.Errorf("plugin validation failed for %s: %w", path, err)
		}
		sink(strings.TrimSuffix(filepath.Base(path), filepath.Ext(path)), o, path, source)
	case json.PluginsSchema:
		var oo Plugins
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		for k := range oo.Plugins {
			plug := oo.Plugins[k]
			if err := plug.Validate(); err != nil {
				return fmt.Errorf("plugin %q validation failed for %s: %w", k, path, err)
			}
			sink(k, plug, path, source)
		}
	case json.PluginMultiSchema:
		var oo plugins
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		for k := range oo {
			plug := oo[k]
			if err := plug.Validate(); err != nil {
				return fmt.Errorf("plugin %q validation failed for %s: %w", k, path, err)
			}
			sink(k, plug, path, source)
		}
	}

	return nil
}

func loadPluginDir(dir string, source PluginSource, sink pluginSink) error {
	if dir == "" {
		return nil
	}
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
			errs = errors.Join(errs, loadPluginFile(path, source, sink))
			return nil
		},
		ErrorCallback: func(osPathname string, err error) godirwalk.ErrorAction {
			slog.Warn("Error at %s: %v - skipping node", slogs.Path, osPathname, slogs.Error, err)
			return godirwalk.SkipNode
		},
	}))

	return errs
}

func (p *Plugins) load(path string) error {
	return loadPluginFile(path, PluginSourceGlobal, func(name string, plugin Plugin, _ string, _ PluginSource) {
		p.Plugins[name] = plugin
	})
}

func (p Plugins) loadDir(dir string) error {
	return loadPluginDir(dir, PluginSourceXDGDataDir, func(name string, plugin Plugin, _ string, _ PluginSource) {
		p.Plugins[name] = plugin
	})
}
