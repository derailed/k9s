// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package data

const DefaultView = "po"

// View tracks view configuration options.
type View struct {
	Active string `yaml:"active"`
}

// NewView creates a new view configuration.
func NewView() *View {
	return &View{Active: DefaultView}
}

// Validate a view configuration.
func (v *View) Validate() {
	if len(v.Active) == 0 {
		v.Active = DefaultView
	}
}
