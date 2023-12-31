// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

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
func (Base) ColorerFunc() ColorerFunc {
	return DefaultColorer
}

// Happy returns true if resource is happy, false otherwise.
func (Base) Happy(_ string, _ Row) bool {
	return true
}
