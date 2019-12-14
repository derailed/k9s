package model

import (
	"context"
	"errors"
	"sort"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// Alias represents a collection of aliases.
type Alias struct {
	Resource
}

// List returns a collection of screen dumps.
func (b *Alias) List(ctx context.Context) ([]runtime.Object, error) {
	aa, ok := ctx.Value(internal.KeyAliases).(config.Alias)
	if !ok {
		return nil, errors.New("no aliases found in context")
	}

	m := make(config.ShortNames, len(aa))
	for alias, gvr := range aa {
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

// Hydrate returns a pod as container rows.
func (b *Alias) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, render.NonResource, &rr[i]); err != nil {
			return err
		}
	}
	return nil
}
