package model

import (
	"context"
	"errors"
	"sort"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// Alias represents a collection of aliases.
type Alias struct {
	Resource
}

// List returns a collection of screen dumps.
func (b *Alias) List(ctx context.Context) ([]runtime.Object, error) {
	a, ok := ctx.Value(internal.KeyAliases).(*dao.Alias)
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
