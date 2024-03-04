// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"github.com/derailed/k9s/internal/model1"
)

// DecoratorFunc decorates a string.
type DecoratorFunc func(string) string

// AgeDecorator represents a timestamped as human column.
var AgeDecorator = func(a string) string {
	return toAgeHuman(a)
}

type Base struct{}

// IsGeneric identifies a generic handler.
func (Base) IsGeneric() bool {
	return false
}

// ColorerFunc colors a resource row.
func (Base) ColorerFunc() model1.ColorerFunc {
	return model1.DefaultColorer
}

// Happy returns true if resource is happy, false otherwise.
func (Base) Happy(string, *model1.Row) bool {
	return true
}
