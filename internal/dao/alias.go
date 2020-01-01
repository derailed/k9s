package dao

import (
	"strings"

	"github.com/derailed/k9s/internal/config"
)

// Alias tracks standard and custom command aliases.
type Alias struct {
	config.Aliases
	factory Factory
}

// NewAlias returns a new set of aliases.
func NewAlias(f Factory) *Alias {
	return &Alias{
		Aliases: config.NewAliases(),
		factory: f,
	}
}

// Clear remove all aliases.
func (a *Alias) Clear() {
	for k := range a.Alias {
		delete(a.Alias, k)
	}
}

// Ensure makes sure alias are loaded.
func (a *Alias) Ensure() (config.Alias, error) {
	// if len(a.Alias) > 0 {
	// 	return a.Alias, nil
	// }
	if err := LoadResources(a.factory); err != nil {
		return config.Alias{}, err
	}
	return a.Alias, a.load()
}

func (a *Alias) load() error {
	if err := a.Load(); err != nil {
		return err
	}

	for _, gvr := range AllGVRs() {
		meta, err := MetaFor(gvr)
		if err != nil {
			return err
		}
		if _, ok := a.Alias[meta.Kind]; ok {
			continue
		}
		a.Define(string(gvr), strings.ToLower(meta.Kind), meta.Name)
		if meta.SingularName != "" {
			a.Define(string(gvr), meta.SingularName)
		}
		if meta.ShortNames != nil {
			a.Define(string(gvr), meta.ShortNames...)
		}
	}

	return nil
}
