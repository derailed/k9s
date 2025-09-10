// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"log/slog"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/slogs"
)

// DecoratorFunc decorates a string.
type DecoratorFunc func(string) string

// AgeDecorator represents a timestamped as human column.
var AgeDecorator = toAgeHuman

type Base struct {
	vs         *config.ViewSetting
	specs      ColumnSpecs
	includeObj bool
}

func (b *Base) SetIncludeObject(f bool) {
	b.includeObj = f
}

// IsGeneric identifies a generic handler.
func (*Base) IsGeneric() bool {
	return false
}

func (b *Base) doHeader(dh model1.Header) model1.Header {
	if b.specs.isEmpty() {
		return dh
	}

	return b.specs.Header(dh)
}

// SetViewSetting sets custom view settings if any.
func (b *Base) SetViewSetting(vs *config.ViewSetting) {
	var cols []string
	b.vs = vs
	if vs != nil {
		cols = vs.Columns
	}
	specs, err := NewColsSpecs(cols...).parseSpecs()
	if err != nil {
		slog.Error("Unable to grok custom columns", slogs.Error, err)
		return
	}
	b.specs = specs
}

// ColorerFunc colors a resource row.
func (*Base) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Happy returns true if resource is happy, false otherwise.
func (*Base) Happy(string, *model1.Row) bool {
	return true
}

// Healthy checks if the resource is healthy.
func (*Base) Healthy(context.Context, any) error {
	return nil
}
