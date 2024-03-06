// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sync"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// Alias tracks shortname to GVR mappings.
type Alias map[string]string

// ShortNames represents a collection of shortnames for aliases.
type ShortNames map[string][]string

// Aliases represents a collection of aliases.
type Aliases struct {
	Alias Alias `yaml:"aliases"`
	mx    sync.RWMutex
}

// NewAliases return a new alias.
func NewAliases() *Aliases {
	return &Aliases{
		Alias: make(Alias, 50),
	}
}

func (a *Aliases) AliasesFor(s string) []string {
	aa := make([]string, 0, 10)

	a.mx.RLock()
	defer a.mx.RUnlock()
	for k, v := range a.Alias {
		if v == s {
			aa = append(aa, k)
		}
	}

	return aa
}

// Keys returns all aliases keys.
func (a *Aliases) Keys() []string {
	a.mx.RLock()
	defer a.mx.RUnlock()

	ss := make([]string, 0, len(a.Alias))
	for k := range a.Alias {
		ss = append(ss, k)
	}
	return ss
}

// ShortNames return all shortnames.
func (a *Aliases) ShortNames() ShortNames {
	a.mx.RLock()
	defer a.mx.RUnlock()

	m := make(ShortNames, len(a.Alias))
	for alias, gvr := range a.Alias {
		if v, ok := m[gvr]; ok {
			m[gvr] = append(v, alias)
		} else {
			m[gvr] = []string{alias}
		}
	}

	return m
}

// Clear remove all aliases.
func (a *Aliases) Clear() {
	a.mx.Lock()
	defer a.mx.Unlock()

	for k := range a.Alias {
		delete(a.Alias, k)
	}
}

// Get retrieves an alias.
func (a *Aliases) Get(k string) (string, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	v, ok := a.Alias[k]
	return v, ok
}

// Define declares a new alias.
func (a *Aliases) Define(gvr string, aliases ...string) {
	a.mx.Lock()
	defer a.mx.Unlock()

	// BOZO!! Could not get full events struct using this api group??
	if gvr == "events.k8s.io/v1/events" || gvr == "extensions/v1beta1" {
		return
	}

	for _, alias := range aliases {
		if _, ok := a.Alias[alias]; ok {
			continue
		}
		a.Alias[alias] = gvr
	}
}

// Load K9s aliases.
func (a *Aliases) Load(path string) error {
	a.loadDefaultAliases()

	f, err := EnsureAliasesCfgFile()
	if err != nil {
		log.Error().Err(err).Msgf("Unable to gen config aliases")
	}

	// load global alias file
	if err := a.LoadFile(f); err != nil {
		return err
	}

	// load context specific aliases if any
	return a.LoadFile(path)
}

// LoadFile loads alias from a given file.
func (a *Aliases) LoadFile(path string) error {
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
	if err := data.JSONValidator.Validate(json.AliasesSchema, bb); err != nil {
		return fmt.Errorf("validation failed for %q: %w", path, err)
	}

	var aa Aliases
	if err := yaml.Unmarshal(bb, &aa); err != nil {
		return err
	}
	a.mx.Lock()
	defer a.mx.Unlock()
	for k, v := range aa.Alias {
		a.Alias[k] = v
	}

	return nil
}

func (a *Aliases) declare(key string, aliases ...string) {
	a.Alias[key] = key
	for _, alias := range aliases {
		a.Alias[alias] = key
	}
}

func (a *Aliases) loadDefaultAliases() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.declare("help", "h", "?")
	a.declare("quit", "q", "q!", "qa", "Q")
	a.declare("aliases", "alias", "a")
	// !!BOZO!!
	// a.declare("popeye", "pop")
	a.declare("helm", "charts", "chart", "hm")
	a.declare("dir", "d")
	a.declare("contexts", "context", "ctx")
	a.declare("users", "user", "usr")
	a.declare("groups", "group", "grp")
	a.declare("portforwards", "portforward", "pf")
	a.declare("benchmarks", "benchmark", "bench")
	a.declare("screendumps", "screendump", "sd")
	a.declare("pulses", "pulse", "pu", "hz")
	a.declare("xrays", "xray", "x")
	a.declare("workloads", "workload", "wk")
}

// Save alias to disk.
func (a *Aliases) Save() error {
	log.Debug().Msg("[Config] Saving Aliases...")
	return a.SaveAliases(AppAliasesFile)
}

// SaveAliases saves aliases to a given file.
func (a *Aliases) SaveAliases(path string) error {
	if err := data.EnsureDirPath(path, data.DefaultDirMod); err != nil {
		return err
	}
	cfg, err := yaml.Marshal(a)
	if err != nil {
		return err
	}

	return os.WriteFile(path, cfg, data.DefaultFileMod)
}
