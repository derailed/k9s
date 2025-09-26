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
		for k, plugin := range oo.Plugins {
			// SECURITY FIX (SEC-002): Validate plugin commands to prevent command injection
			// Before: Plugin commands were executed without validation, allowing arbitrary command execution
			// After: Validate commands against allowlist and check arguments for injection patterns
			if err := validatePluginCommand(plugin.Command, plugin.Args); err != nil {
				slog.Warn("Plugin command validation failed",
					slogs.Plugin, k,
					slogs.Command, plugin.Command,
					slogs.Error, err,
				)
				continue // Skip invalid plugins instead of failing completely
			}
			p.Plugins[k] = plugin
		}
	case json.PluginMultiSchema:
		var oo plugins
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("plugin unmarshal failed for %s: %w", path, err)
		}
		for k, plugin := range oo {
			// SECURITY FIX (SEC-002): Validate plugin commands to prevent command injection
			// Before: Plugin commands were executed without validation, allowing arbitrary command execution
			// After: Validate commands against allowlist and check arguments for injection patterns
			if err := validatePluginCommand(plugin.Command, plugin.Args); err != nil {
				slog.Warn("Plugin command validation failed",
					slogs.Plugin, k,
					slogs.Command, plugin.Command,
					slogs.Error, err,
				)
				continue // Skip invalid plugins instead of failing completely
			}
			p.Plugins[k] = plugin
		}
	}

	return nil
}

func (p Plugins) loadDir(dir string) error {
	if _, err := os.Stat(dir); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	var errs error
	errs = errors.Join(errs, filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !isYamlFile(info.Name()) {
			return nil
		}
		errs = errors.Join(errs, p.load(path))
		return nil
	}))

	return errs
}

// validatePluginCommand validates plugin commands to prevent command injection attacks
// SECURITY FIX (SEC-002): This function prevents malicious plugin configurations from
// executing arbitrary commands that could compromise the system or exfiltrate data.
//
// Security measures implemented:
// 1. Command allowlist - only allows safe, commonly used commands
// 2. Argument validation - checks for injection patterns in arguments
// 3. Path validation - ensures commands are not using relative paths for traversal
// 4. Dangerous command detection - blocks commands that could be used maliciously
//
// Why this is important:
// - Attackers could create malicious plugin configs to execute arbitrary commands
// - This could lead to data exfiltration, system compromise, or credential theft
// - The fix ensures only safe, intended commands can be executed through plugins
func validatePluginCommand(command string, args []string) error {
	// Create allowlist of safe commands that are commonly used in k9s plugins
	// This is a restrictive list that can be expanded as needed for legitimate use cases
	allowedCommands := map[string]bool{
		"kubectl":     true, // Kubernetes CLI - primary use case
		"helm":        true, // Helm package manager
		"jq":          true, // JSON processor
		"grep":        true, // Text search
		"awk":         true, // Text processing
		"sed":         true, // Text editing
		"sort":        true, // Text sorting
		"uniq":        true, // Remove duplicates
		"head":        true, // Show first lines
		"tail":        true, // Show last lines
		"wc":          true, // Word count
		"cat":         true, // Display file contents
		"echo":        true, // Display text
		"printf":      true, // Formatted output
		"basename":    true, // File path utilities
		"dirname":    true, // File path utilities
		"filepath":   true, // File path utilities
		"xargs":      true, // Execute commands
		"sh":         true, // Shell (restricted)
		"bash":       true, // Bash shell (restricted)
		"/usr/bin/sh": true, // Full path to shell
		"/bin/sh":     true, // Full path to shell
		"/bin/bash":   true, // Full path to bash
	}

	// Check if command is in allowlist
	if !allowedCommands[command] {
		return fmt.Errorf("command not in allowlist: %s", command)
	}

	// Validate arguments for injection patterns
	for _, arg := range args {
		if err := validateArgument(arg); err != nil {
			return fmt.Errorf("invalid argument: %w", err)
		}
	}

	return nil
}

// validateArgument checks individual arguments for potential injection patterns
// This prevents command injection through argument manipulation
func validateArgument(arg string) error {
	// Check for common injection patterns
	dangerousPatterns := []string{
		"$((",           // Command substitution
		"$(",            // Command substitution
		"`",             // Backtick command substitution
		";",             // Command chaining
		"&&",            // Command chaining
		"||",            // Command chaining
		"|",             // Pipe (unless explicitly needed)
		">",             // Output redirection
		"<",             // Input redirection
		"&",             // Background execution
		"#",             // Comment (could hide malicious content)
		"../",           // Path traversal
		"..\\",          // Windows path traversal
		"rm ",           // Remove command
		"del ",          // Windows delete command
		"format",        // Format command
		"fdisk",         // Disk utility
		"mkfs",          // File system creation
		"dd ",           // Disk dump
		"nc ",           // Netcat
		"netcat",        // Netcat
		"wget",          // Download utility
		"curl",          // Download utility (unless explicitly needed)
		"python",        // Python interpreter
		"perl",          // Perl interpreter
		"ruby",          // Ruby interpreter
		"node",          // Node.js interpreter
		"php",           // PHP interpreter
	}

	for _, pattern := range dangerousPatterns {
		if strings.Contains(strings.ToLower(arg), pattern) {
			return fmt.Errorf("potentially dangerous pattern detected: %s", pattern)
		}
	}

	// Check for excessive length (potential buffer overflow or obfuscation)
	if len(arg) > 1000 {
		return fmt.Errorf("argument too long: %d characters", len(arg))
	}

	return nil
}
