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
		a.Define(gvr.String(), strings.ToLower(meta.Kind), meta.Name)
		if meta.SingularName != "" {
			a.Define(gvr.String(), meta.SingularName)
		}
		if meta.ShortNames != nil {
			a.Define(gvr.String(), meta.ShortNames...)
		}
	}

	return nil
}
