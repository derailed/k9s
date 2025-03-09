// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v2"
)

// HotKeys represents a collection of plugins.
type HotKeys struct {
	HotKey map[string]HotKey `yaml:"hotKeys"`
}

// HotKey describes a K9s hotkey.
type HotKey struct {
	ShortCut    string `yaml:"shortCut"`
	Override    bool   `yaml:"override"`
	Description string `yaml:"description"`
	Command     string `yaml:"command"`
	KeepHistory bool   `yaml:"keepHistory"`
}

// NewHotKeys returns a new plugin.
func NewHotKeys() HotKeys {
	return HotKeys{
		HotKey: make(map[string]HotKey),
	}
}

// Load K9s plugins.
func (h HotKeys) Load(path string) error {
	if err := h.LoadHotKeys(AppHotKeysFile); err != nil {
		return err
	}
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	return h.LoadHotKeys(path)
}

// LoadHotKeys loads plugins from a given file.
func (h HotKeys) LoadHotKeys(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.HotkeysSchema, bb); err != nil {
		slog.Warn("Validation failed. Please update your config and restart.",
			slogs.Path, path,
			slogs.Error, err,
		)
	}

	var hh HotKeys
	if err := yaml.Unmarshal(bb, &hh); err != nil {
		return err
	}
	for k, v := range hh.HotKey {
		h.HotKey[k] = v
	}

	return nil
}
