package dao

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Alias)(nil)

// Alias tracks standard and custom command aliases.
type Alias struct {
	NonResource
	config.Aliases
}

// NewAlias returns a new set of aliases.
func NewAlias(f Factory) *Alias {
	a := Alias{Aliases: config.NewAliases()}
	a.Init(f, client.NewGVR("aliases"))

	return &a
}

// Clear remove all aliases.
func (a *Alias) Clear() {
	for k := range a.Alias {
		delete(a.Alias, k)
	}
}

// List returns a collection of screen dumps.
// BOZO!! Already have aliases here. Refact!!
func (a *Alias) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	a, ok := ctx.Value(internal.KeyAliases).(*Alias)
	if !ok {
		return nil, errors.New("no aliases found in context")
	}

	m := make(config.ShortNames, len(a.Alias))
	for alias, gvr := range a.Alias {
		if _, ok := m[gvr]; ok {
			m[gvr] = append(m[gvr], alias)
		} else {
			m[gvr] = []string{alias}
		}
	}

	oo := make([]runtime.Object, 0, len(m))
	for gvr, aliases := range m {
		sort.StringSlice(aliases).Sort()
		oo = append(oo, render.AliasRes{GVR: gvr, Aliases: aliases})
	}

	return oo, nil
}

// AsGVR returns a matching gvr if it exists.
func (a *Alias) AsGVR(cmd string) (client.GVR, bool) {
	gvr, ok := a.Aliases.Get(cmd)
	if ok {
		return client.NewGVR(gvr), true
	}
	return client.GVR{}, false
}

// Get fetch a resource.
func (a *Alias) Get(_ context.Context, _ string) (runtime.Object, error) {
	// BOZO!! NYI
	panic("NYI!")
}

// Ensure makes sure alias are loaded.
func (a *Alias) Ensure() (config.Alias, error) {
	if err := LoadResources(a.Factory); err != nil {
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
		if _, ok := a.Alias[meta.Kind]; ok || IsK9sMeta(meta) {
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
