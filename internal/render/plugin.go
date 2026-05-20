// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const pluginRowIDSep = "\x1f"

// Plugin renders a K9s plugin catalog entry.
type Plugin struct{}

// IsGeneric identifies a generic handler.
func (Plugin) IsGeneric() bool {
	return false
}

// Healthy checks if the resource is healthy.
func (Plugin) Healthy(context.Context, any) error {
	return nil
}

// ColorerFunc colors a resource row.
func (Plugin) ColorerFunc() model1.ColorerFunc {
	return func(string, model1.Header, *model1.RowEvent) tcell.Color {
		return tcell.ColorLightSeaGreen
	}
}

func (Plugin) SetViewSetting(*config.ViewSetting) {}

// Header returns a header row.
func (Plugin) Header(string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "SHORTCUT"},
		model1.HeaderColumn{Name: "SCOPES"},
		model1.HeaderColumn{Name: "LOCATION"},
		model1.HeaderColumn{Name: "COMMAND"},
		model1.HeaderColumn{Name: "SOURCE", Attrs: model1.Attrs{Wide: true}},
	}
}

// Render renders a plugin catalog entry to screen.
func (Plugin) Render(o any, _ string, r *model1.Row) error {
	p, ok := o.(PluginRes)
	if !ok {
		return fmt.Errorf("expected PluginRes, but got %T", o)
	}

	r.ID = PluginRowID(p.Path, p.Name)
	r.Fields = append(r.Fields,
		p.Name,
		p.Plugin.ShortCut,
		strings.Join(p.Plugin.Scopes, ", "),
		string(p.Source),
		p.Plugin.Command,
		p.Path,
	)

	return nil
}

// PluginRowID builds a stable row identifier.
func PluginRowID(path, name string) string {
	return path + pluginRowIDSep + name
}

// ParsePluginRowID extracts the file path and plugin name from a row id.
func ParsePluginRowID(id string) (path, name string) {
	path, name, _ = strings.Cut(id, pluginRowIDSep)
	return path, name
}

// PluginRes represents an effective K9s plugin.
type PluginRes struct {
	Name   string
	Path   string
	Source config.PluginSource
	Plugin config.Plugin
}

// GetObjectKind returns a schema object.
func (PluginRes) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (p PluginRes) DeepCopyObject() runtime.Object {
	return p
}
