package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// K9sAlias stores K9s command aliases.
var K9sAlias = filepath.Join(K9sHome, "alias.yml")

type Alias map[string]string

// Aliases represents a collection of aliases.
type Aliases struct {
	Alias Alias `yaml:"alias"`
}

// NewAliases return a new alias.
func NewAliases() Aliases {
	aa := Aliases{Alias: make(Alias, 50)}
	aa.loadDefaults()
	return aa
}

func (a Aliases) loadDefaults() {
	a.Alias["dp"] = "apps/v1/deployments"
	a.Alias["sec"] = "v1/secrets"
	a.Alias["jo"] = "batch/v1/jobs"
	a.Alias["cr"] = "rbac.authorization.k8s.io/v1/clusterroles"
	a.Alias["crb"] = "rbac.authorization.k8s.io/v1/clusterrolebindings"
	a.Alias["ro"] = "rbac.authorization.k8s.io/v1/roles"
	a.Alias["rob"] = "rbac.authorization.k8s.io/v1/rolebindings"
	a.Alias["np"] = "networking.k8s.io/v1beta1/rolebindings"
	{
		a.Alias["ctx"] = "contexts"
		a.Alias["contexts"] = "contexts"
		a.Alias["context"] = "contexts"
	}
	{
		a.Alias["usr"] = "users"
		a.Alias["users"] = "users"
		a.Alias["user"] = "user"
	}
	{
		a.Alias["grp"] = "groups"
		a.Alias["group"] = "groups"
		a.Alias["groups"] = "groups"
	}
	{
		a.Alias["pf"] = "portforwards"
		a.Alias["portforwards"] = "portforwards"
		a.Alias["portforward"] = "portforwards"
	}
	{
		a.Alias["be"] = "benchmarks"
		a.Alias["benchmark"] = "benchmarks"
		a.Alias["benchmarks"] = "benchmarks"
	}
	{
		a.Alias["sd"] = "screendumps"
		a.Alias["screendump"] = "screendumps"
		a.Alias["screendumps"] = "screendumps"
	}
}

// Load K9s aliases.
func (a Aliases) Load() error {
	return a.LoadAliases(K9sAlias)
}

// Get retrieves an alias.
func (a Aliases) Get(k string) (string, bool) {
	v, ok := a.Alias[k]
	return v, ok
}

// Define declares a new alias.
func (a Aliases) Define(args ...string) {
	if len(args)%2 != 0 {
		panic("Invalid alias definition. You must specify pairs")
	}
	for i := 0; i < len(args); i += 2 {
		a.Alias[args[i]] = args[i+1]
	}
}

// LoadAliases K9s alias from a given file.
func (a Aliases) LoadAliases(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var aa Aliases
	if err := yaml.Unmarshal(f, &aa); err != nil {
		return err
	}
	for k, v := range aa.Alias {
		a.Alias[k] = v
	}

	return nil
}

// Save alias to disk.
func (a Aliases) Save() error {
	log.Debug().Msg("[Config] Saving Aliases...")
	return a.SaveAliases(K9sAlias)
}

// SaveAliases saves aliases to a given file.
func (a Aliases) SaveAliases(path string) error {
	EnsurePath(path, DefaultDirMod)
	cfg, err := yaml.Marshal(a)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, cfg, 0644)
}
