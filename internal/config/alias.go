// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/view/cmd"
	"gopkg.in/yaml.v3"
	"k8s.io/apimachinery/pkg/util/sets"
)

type (
	// Alias tracks shortname to GVR mappings.
	Alias map[string]*client.GVR

	// ShortNames represents a collection of shortnames for aliases.
	ShortNames map[*client.GVR][]string

	// Aliases represents a collection of aliases.
	Aliases struct {
		Alias Alias `yaml:"aliases"`
		mx    sync.RWMutex
	}
)

// NewAliases return a new alias.
func NewAliases() *Aliases {
	return &Aliases{
		Alias: make(Alias, 50),
	}
}

func (a *Aliases) AliasesFor(gvr *client.GVR) sets.Set[string] {
	a.mx.RLock()
	defer a.mx.RUnlock()

	ss := sets.New[string]()
	for alias, aliasGVR := range a.Alias {
		if aliasGVR == gvr {
			ss.Insert(alias)
		}
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

func (a *Aliases) Resolve(command string) (*client.GVR, string, bool) {
	agvr, ok := a.Get(command)
	if !ok {
		return nil, "", false
	}

	p := cmd.NewInterpreter(agvr.String())
	gvr, ok := a.Get(p.Cmd())
	if !ok {
		return agvr, "", true
	}

	return gvr, p.Args(), true
}

// Get retrieves an alias.
func (a *Aliases) Get(alias string) (*client.GVR, bool) {
	a.mx.RLock()
	defer a.mx.RUnlock()

	gvr, ok := a.Alias[alias]

	return gvr, ok
}

// Define declares a new alias.
func (a *Aliases) Define(gvr *client.GVR, aliases ...string) {
	a.mx.Lock()
	defer a.mx.Unlock()

	for _, alias := range aliases {
		if _, ok := a.Alias[alias]; !ok && alias != "" {
			a.Alias[alias] = gvr
		}
	}
}

// Load K9s aliases.
func (a *Aliases) Load(path string) error {
	a.loadDefaultAliases()
	f, err := EnsureAliasesCfgFile()
	if err != nil {
		slog.Error("Unable to gen config aliases", slogs.Error, err)
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
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}

	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.AliasesSchema, bb); err != nil {
		slog.Warn("Aliases validation failed", slogs.Error, err)
	}

	a.mx.Lock()
	defer a.mx.Unlock()
	if err := yaml.Unmarshal(bb, a); err != nil {
		return err
	}

	return nil
}

func (a *Aliases) declare(gvr *client.GVR, aliases ...string) {
	a.Alias[gvr.String()] = gvr
	for _, alias := range aliases {
		a.Alias[alias] = gvr
	}
}

func (a *Aliases) loadDefaultAliases() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.declare(client.HlpGVR, "h", "?")
	a.declare(client.QGVR, "q", "q!", "qa", "Q")
	a.declare(client.AliGVR, "alias", "a")
	a.declare(client.HmGVR, "charts", "chart", "hm")
	a.declare(client.DirGVR, "dir", "d")
	a.declare(client.CtGVR, "context", "ctx")
	a.declare(client.UsrGVR, "user", "usr")
	a.declare(client.GrpGVR, "group", "grp")
	a.declare(client.PfGVR, "portforward", "pf")
	a.declare(client.BeGVR, "benchmark", "bench")
	a.declare(client.SdGVR, "screendump", "sd")
	a.declare(client.PuGVR, "pulse", "pu", "hz")
	a.declare(client.XGVR, "xray", "x")
	a.declare(client.WkGVR, "workload", "wk")
}

// Save alias to disk.
func (a *Aliases) Save() error {
	slog.Debug("Saving Aliases...")
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.saveAliases(AppAliasesFile)
}

// SaveAliases saves aliases to a given file.
func (a *Aliases) saveAliases(path string) error {
	if err := data.EnsureDirPath(path, data.DefaultDirMod); err != nil {
		return err
	}

	return data.SaveYAML(path, a)
}
