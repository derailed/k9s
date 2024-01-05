// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"

	"gopkg.in/yaml.v2"
)

// HotKeys represents a collection of plugins.
type HotKeys struct {
	HotKey map[string]HotKey `yaml:"hotKeys"`
}

// HotKey describes a K9s hotkey.
type HotKey struct {
	ShortCut    string `yaml:"shortCut"`
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
func (h HotKeys) Load() error {
	return h.LoadHotKeys(AppHotKeysFile)
}

// LoadHotKeys loads plugins from a given file.
func (h HotKeys) LoadHotKeys(path string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	var hh HotKeys
	if err := yaml.Unmarshal(f, &hh); err != nil {
		return err
	}
	for k, v := range hh.HotKey {
		h.HotKey[k] = v
	}

	return nil
}
