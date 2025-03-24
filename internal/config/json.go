// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v3"
)

type JsonTemplate struct {
	Name               string `json:"name" yaml:"name"`
	LogLevelExpression string `json:"loglevel" yaml:"loglevel"`
	DateTimeExpression string `json:"datetime" yaml:"datetime"`
	MessageExpression  string `json:"message" yaml:"message"`
}

type JsonConfig struct {
	Debug             bool           `json:"debug" yaml:"debug"`
	GlobalExpressions string         `json:"globalExpressions" yaml:"globalExpressions"`
	DefaultTemplate   string         `json:"defaultTemplate" yaml:"defaultTemplate"`
	Templates         []JsonTemplate `json:"templates" yaml:"templates"`
}

// Json represents a json config, including templates.
type Json struct {
	JsonConfig JsonConfig `yaml:"json"`
	mx         sync.RWMutex
}

// NewJsonConfig returns a new instance.
func newJsonConfig() JsonConfig {
	return JsonConfig{
		Debug:             false,
		GlobalExpressions: "",
		DefaultTemplate:   "",
		Templates:         []JsonTemplate{},
	}
}

// NewJson return a new json config wrapper.
func NewJson() *Json {
	return &Json{
		JsonConfig: newJsonConfig(),
	}
}

// Load K9s aliases.
func (a *Json) Load(path string) error {
	a.loadDefaultJsonConfig()

	f, err := EnsureJsonCfgFile()
	if err != nil {
		slog.Error("Unable to gen config aliases", slogs.Error, err)
	}

	// load global json config file
	if err := a.LoadFile(f); err != nil {
		return err
	}

	// load context specific json config if any
	return a.LoadFile(path)
}

// LoadFile loads json config from a given file.
func (a *Json) LoadFile(path string) error {
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
	if err := data.JSONValidator.Validate(json.JsonSchema, bb); err != nil {
		slog.Warn("Json validation failed", slogs.Error, err)
	}

	var aa Json
	if err := yaml.Unmarshal(bb, &aa); err != nil {
		return err
	}
	a.mx.Lock()
	defer a.mx.Unlock()

	a.JsonConfig = aa.JsonConfig

	return nil
}

func (a *Json) loadDefaultJsonConfig() {
	a.mx.Lock()
	defer a.mx.Unlock()

	// TODO: default config
}

// Save json config to disk.
func (a *Json) Save() error {
	slog.Debug("Saving Aliases...")
	return a.SaveJsonConfig(AppJsonFile)
}

// SaveJsonConfig saves aliases to a given file.
func (a *Json) SaveJsonConfig(path string) error {
	if err := data.EnsureDirPath(path, data.DefaultDirMod); err != nil {
		return err
	}

	return data.SaveYAML(path, a)
}
