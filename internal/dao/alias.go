package dao

import (
	"context"
	"errors"
	"fmt"
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
	*config.Aliases
}

// NewAlias returns a new set of aliases.
func NewAlias(f Factory) *Alias {
	a := Alias{Aliases: config.NewAliases()}
	a.Init(f, client.NewGVR("aliases"))

	return &a
}

// Check verifies an alias is defined for this command.
func (a *Alias) Check(cmd string) bool {
	_, ok := a.Aliases.Get(cmd)
	return ok
}

// List returns a collection of aliases.
func (a *Alias) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	aa, ok := ctx.Value(internal.KeyAliases).(*Alias)
	if !ok {
		return nil, fmt.Errorf("expecting *Alias but got %T", ctx.Value(internal.KeyAliases))
	}
	m := aa.ShortNames()
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
	return nil, errors.New("NYI!!")
}

// Ensure makes sure alias are loaded.
func (a *Alias) Ensure() (config.Alias, error) {
	if err := MetaAccess.LoadResources(a.Factory); err != nil {
		return config.Alias{}, err
	}
	return a.Alias, a.load()
}

func (a *Alias) load() error {
	if err := a.Load(); err != nil {
		return err
	}

	crdGVRS := make(client.GVRs, 0, 50)
	for _, gvr := range MetaAccess.AllGVRs() {
		meta, err := MetaAccess.MetaFor(gvr)
		if err != nil {
			return err
		}
		if IsK9sMeta(meta) {
			continue
		}

		gvrStr := gvr.String()
		if IsCRD(meta) {
			crdGVRS = append(crdGVRS, gvr)
			continue
		}

		a.Define(gvrStr, strings.ToLower(meta.Kind), meta.Name)
		if meta.SingularName != "" {
			a.Define(gvrStr, meta.SingularName)
		}
		if meta.ShortNames != nil {
			a.Define(gvrStr, meta.ShortNames...)
		}
		a.Define(gvrStr, gvrStr)
	}

	for _, gvr := range crdGVRS {
		meta, err := MetaAccess.MetaFor(gvr)
		if err != nil {
			return err
		}
		gvrStr := gvr.String()
		a.Define(gvrStr, strings.ToLower(meta.Kind), meta.Name)
		if meta.SingularName != "" {
			a.Define(gvrStr, meta.SingularName)
		}
		if meta.ShortNames != nil {
			a.Define(gvrStr, meta.ShortNames...)
		}
		a.Define(gvrStr, gvrStr)
		a.Define(gvrStr, meta.Name+"."+meta.Group)
	}

	return nil
}
