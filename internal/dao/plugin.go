// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"sort"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Plugin)(nil)

// Plugin lists the effective K9s plugins plus their source files.
type Plugin struct {
	NonResource
}

// NewPlugin returns a new plugin accessor.
func NewPlugin(f Factory) *Plugin {
	var p Plugin
	p.Init(f, client.PlgGVR)
	return &p
}

// List returns the effective plugin catalog.
func (*Plugin) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	path, _ := ctx.Value(internal.KeyPath).(string)

	catalog := config.NewPluginCatalog()
	if err := catalog.Load(path, true); err != nil {
		return nil, err
	}

	names := make([]string, 0, len(catalog.Entries))
	for name := range catalog.Entries {
		names = append(names, name)
	}
	sort.Strings(names)

	oo := make([]runtime.Object, 0, len(names))
	for _, name := range names {
		entry := catalog.Entries[name]
		oo = append(oo, render.PluginRes{
			Name:   entry.Name,
			Path:   entry.Path,
			Source: entry.Source,
			Plugin: entry.Plugin,
		})
	}

	return oo, nil
}
