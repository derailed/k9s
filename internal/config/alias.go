package config

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// K9sAlias manages K9s aliases.
var K9sAlias = filepath.Join(K9sHome(), "alias.yml")

// Alias tracks shortname to GVR mappings.
type Alias map[string]string

// ShortNames represents a collection of shortnames for aliases.
type ShortNames map[string][]string

// Aliases represents a collection of aliases.
type Aliases struct {
	Alias Alias `yaml:"alias"`
	mx    sync.RWMutex
}

// NewAliases return a new alias.
func NewAliases() *Aliases {
	return &Aliases{
		Alias: make(Alias, 50),
	}
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
func (a *Aliases) Load() error {
	a.loadDefaultAliases()
	return a.LoadFileAliases(K9sAlias)
}

// LoadFileAliases loads alias from a given file.
func (a *Aliases) LoadFileAliases(path string) error {
	f, err := os.ReadFile(path)
	if err == nil {
		var aa Aliases
		if err := yaml.Unmarshal(f, &aa); err != nil {
			return err
		}

		a.mx.Lock()
		defer a.mx.Unlock()
		for k, v := range aa.Alias {
			a.Alias[k] = v
		}
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

	a.Alias["dp"] = "apps/v1/deployments"
	a.Alias["sec"] = "v1/secrets"
	a.Alias["jo"] = "batch/v1/jobs"
	a.Alias["cr"] = "rbac.authorization.k8s.io/v1/clusterroles"
	a.Alias["crb"] = "rbac.authorization.k8s.io/v1/clusterrolebindings"
	a.Alias["ro"] = "rbac.authorization.k8s.io/v1/roles"
	a.Alias["rb"] = "rbac.authorization.k8s.io/v1/rolebindings"
	a.Alias["np"] = "networking.k8s.io/v1/networkpolicies"

	a.declare("help", "h", "?")
	a.declare("quit", "q", "Q")
	a.declare("aliases", "alias", "a")
	a.declare("popeye", "pop")
	a.declare("helm", "charts", "chart", "hm")
	a.declare("dir", "d")
	a.declare("contexts", "context", "ctx")
	a.declare("users", "user", "usr")
	a.declare("groups", "group", "grp")
	a.declare("portforwards", "portforward", "pf")
	a.declare("benchmarks", "bench", "benchmark", "be")
	a.declare("screendumps", "screendump", "sd")
	a.declare("pulses", "pulse", "pu", "hz")
	a.declare("xrays", "xray", "x")
}

// Save alias to disk.
func (a *Aliases) Save() error {
	log.Debug().Msg("[Config] Saving Aliases...")
	return a.SaveAliases(K9sAlias)
}

// SaveAliases saves aliases to a given file.
func (a *Aliases) SaveAliases(path string) error {
	EnsurePath(path, DefaultDirMod)
	cfg, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	return os.WriteFile(path, cfg, 0644)
}
