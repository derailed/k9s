package config

import (
	"io/ioutil"
	"path/filepath"
	"sync"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// K9sAlias manages K9s aliases.
var K9sAlias = filepath.Join(K9sHome, "alias.yml")

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

// ShortNames return all shortnames.
func (a *Aliases) ShortNames() ShortNames {
	a.mx.RLock()
	defer a.mx.RUnlock()

	m := make(ShortNames, len(a.Alias))
	for alias, gvr := range a.Alias {
		if _, ok := m[gvr]; ok {
			m[gvr] = append(m[gvr], alias)
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
	f, err := ioutil.ReadFile(path)
	if err != nil {
		log.Debug().Err(err).Msgf("No custom aliases found")
		return nil
	}

	var aa Aliases
	if err := yaml.Unmarshal(f, &aa); err != nil {
		return err
	}

	a.mx.Lock()
	defer a.mx.Unlock()
	for k, v := range aa.Alias {
		a.Alias[k] = v
	}

	return nil
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

	const contexts = "contexts"
	{
		a.Alias["ctx"] = contexts
		a.Alias[contexts] = contexts
		a.Alias["context"] = contexts
	}
	const users = "users"
	{
		a.Alias["usr"] = users
		a.Alias[users] = users
		a.Alias["user"] = users
	}
	const groups = "groups"
	{
		a.Alias["grp"] = groups
		a.Alias["group"] = groups
		a.Alias[groups] = groups
	}
	const portFwds = "portforwards"
	{
		a.Alias["pf"] = portFwds
		a.Alias[portFwds] = portFwds
		a.Alias["portforward"] = portFwds
	}
	const benchmarks = "benchmarks"
	{
		a.Alias["be"] = benchmarks
		a.Alias["benchmark"] = benchmarks
		a.Alias[benchmarks] = benchmarks
	}
	const dumps = "screendumps"
	{
		a.Alias["sd"] = dumps
		a.Alias["screendump"] = dumps
		a.Alias[dumps] = dumps
	}
	const pulses = "pulses"
	{
		a.Alias["hz"] = pulses
		a.Alias["pu"] = pulses
		a.Alias["pulse"] = pulses
	}
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
	return ioutil.WriteFile(path, cfg, 0644)
}
