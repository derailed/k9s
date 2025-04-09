// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

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
	"k8s.io/apimachinery/pkg/util/sets"
)

var _ Accessor = (*Alias)(nil)

// Alias tracks standard and custom command aliases.
type Alias struct {
	NonResource

	*config.Aliases
}

// NewAlias returns a new set of aliases.
func NewAlias(f Factory) *Alias {
	a := Alias{
		Aliases: config.NewAliases(),
	}
	a.Init(f, client.AliGVR)

	return &a
}

// AliasesFor returns a set of aliases for a given gvr.
func (a *Alias) AliasesFor(gvr *client.GVR) sets.Set[string] {
	return a.Aliases.AliasesFor(gvr)
}

// List returns a collection of aliases.
func (*Alias) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	aa, ok := ctx.Value(internal.KeyAliases).(*Alias)
	if !ok {
		return nil, fmt.Errorf("expecting *Alias but got %T", ctx.Value(internal.KeyAliases))
	}
	m := aa.ShortNames()
	oo := make([]runtime.Object, 0, len(m))
	for gvr, aliases := range m {
		sort.StringSlice(aliases).Sort()
		oo = append(oo, render.AliasRes{
			GVR:     gvr,
			Aliases: aliases,
		})
	}

	return oo, nil
}

// AsGVR returns a matching gvr if it exists.
func (a *Alias) AsGVR(alias string) (*client.GVR, string, bool) {
	gvr, ok := a.Aliases.Get(alias)
	if ok {
		if pgvr := MetaAccess.Lookup(alias); pgvr != client.NoGVR {
			return pgvr, "", ok
		}
	}

	return gvr, "", ok
}

// Get fetch a resource.
func (*Alias) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, errors.New("nyi")
}

// Ensure makes sure alias are loaded.
func (a *Alias) Ensure(path string) (config.Alias, error) {
	if err := MetaAccess.LoadResources(a.Factory); err != nil {
		return config.Alias{}, err
	}
	return a.Alias, a.load(path)
}

func (a *Alias) load(path string) error {
	if err := a.Load(path); err != nil {
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
		if IsCRD(meta) {
			crdGVRS = append(crdGVRS, gvr)
			continue
		}
		a.Define(gvr, gvr.AsResourceName())

		// Allow single shot commands for k8s resources only!
		if isStandardGroup(gvr.GVSub()) {
			a.Define(gvr, meta.Name)
			a.Define(gvr, meta.SingularName)
		}
		if len(meta.ShortNames) > 0 {
			a.Define(gvr, meta.ShortNames...)
		}
		a.Define(gvr, gvr.String())
	}

	for _, gvr := range crdGVRS {
		meta, err := MetaAccess.MetaFor(gvr)
		if err != nil {
			return err
		}
		a.Define(gvr, strings.ToLower(meta.Kind), meta.Name)
		a.Define(gvr, meta.SingularName)

		if len(meta.ShortNames) > 0 {
			a.Define(gvr, meta.ShortNames...)
		}
		a.Define(gvr, gvr.String())
		a.Define(gvr, meta.Name+"."+meta.Group)
	}

	return nil
}
