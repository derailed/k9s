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

type jumpers map[string]Jumper

// Jumpers represents a collection of jumpers.
type Jumpers struct {
	Jumpers jumpers `yaml:"jumpers"`
}

// Jumper describes a K9s jumper.
type Jumper struct {
	Variables   []Variable `yaml:"variables"`
	ShortCut    string     `yaml:"shortCut"`
	Override    bool       `yaml:"override"`
	Description string     `yaml:"description"`
	Scopes      []string   `yaml:"scopes"`
	View        string     `yaml:"view"`
	Filters     []string   `yaml:"filters"`
	Labels      []string   `yaml:"labels"`
}

func (p Jumper) String() string {
	return fmt.Sprintf("[%s] %s", p.ShortCut, p.Description)
}

// NewJumpers returns a new jumper.
func NewJumpers() Jumpers {
	return Jumpers{
		Jumpers: make(map[string]Jumper),
	}
}

// Load K9s jumpers.
func (p Jumpers) Load(path string, loadExtra bool) error {
	var errs error

	// Load from global config file
	if err := p.load(AppJumpersFile); err != nil {
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
	const k9sJumpersDir = "k9s/jumpers"
	for _, dir := range append(xdg.DataDirs, xdg.DataHome, xdg.ConfigHome) {
		path := filepath.Join(dir, k9sJumpersDir)
		if err := p.loadDir(path); err != nil {
			errs = errors.Join(errs, err)
		}
	}

	return errs
}

func (p *Jumpers) load(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	scheme, err := data.JSONValidator.ValidateJumpers(bb)
	if err != nil {
		slog.Warn("Jumper schema validation failed",
			slogs.Path, path,
			slogs.Error, err,
		)
		return fmt.Errorf("jumper validation failed for %s: %w", path, err)
	}

	d := yaml.NewDecoder(bytes.NewReader(bb))
	d.KnownFields(true)

	switch scheme {
	case json.JumperSchema:
		var o Jumper
		if err := yaml.Unmarshal(bb, &o); err != nil {
			return fmt.Errorf("jumper unmarshal failed for %s: %w", path, err)
		}
		p.Jumpers[strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))] = o
	case json.JumpersSchema:
		var oo Jumpers
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("jumper unmarshal failed for %s: %w", path, err)
		}
		for k := range oo.Jumpers {
			p.Jumpers[k] = oo.Jumpers[k]
		}
	case json.JumperMultiSchema:
		var oo jumpers
		if err := yaml.Unmarshal(bb, &oo); err != nil {
			return fmt.Errorf("jumper unmarshal failed for %s: %w", path, err)
		}
		for k := range oo {
			p.Jumpers[k] = oo[k]
		}
	}

	return nil
}

func (p Jumpers) loadDir(dir string) error {
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
