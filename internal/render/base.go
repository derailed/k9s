// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/model1"
	"github.com/rs/zerolog/log"
)

// DecoratorFunc decorates a string.
type DecoratorFunc func(string) string

// AgeDecorator represents a timestamped as human column.
var AgeDecorator = func(a string) string {
	return toAgeHuman(a)
}

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

func (b *Base) SetViewSetting(vs *config.ViewSetting) {
	var cols []string
	b.vs = vs
	if vs != nil {
		cols = vs.Columns
	}
	specs, err := NewColsSpecs(cols...).parseSpecs()
	if err != nil {
		log.Error().Err(err).Msg("unable to grok custom columns")
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
