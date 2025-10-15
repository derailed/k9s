// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Variable describes a K9s jumper variable.
type Variable struct {
	Name    string          `yaml:"name"`
	Source  VariableSource  `yaml:"source"`
	Command string          `yaml:"command"`
	Args    []string        `yaml:"args"`
	Pipes   []string        `yaml:"pipes"`
	Display VariableDisplay `yaml:"display"`
}

type VariableSource string

// Jumper variables sources.
const (
	VariableSourceScript VariableSource = "script"
	VariableSourceStatic VariableSource = "static"
)

// Unmarshalling YAML values for the VariableSource type
func (d *VariableSource) UnmarshalYAML(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		return fmt.Errorf("invalid jumper variable source")
	}

	var raw string
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch VariableSource(raw) {
	case VariableSourceStatic,
		VariableSourceScript:
		*d = VariableSource(raw)
	default:
		return fmt.Errorf("unmanaged jumper variable source")
	}
	return nil
}

type VariableDisplay string

// Jumper variables display options.
const (
	VariableDisplayNone   VariableDisplay = "none"
	VariableDisplayText   VariableDisplay = "text"
	VariableDisplaySelect VariableDisplay = "select"
)

// Unmarshalling YAML values for the VariableDisplay type
func (d *VariableDisplay) UnmarshalYAML(b []byte) error {
	if string(b) == "null" || string(b) == `""` {
		*d = VariableDisplayNone
		return nil
	}

	var raw string
	if err := yaml.Unmarshal(b, &raw); err != nil {
		return err
	}
	switch VariableDisplay(raw) {
	case VariableDisplayNone,
		VariableDisplayText,
		VariableDisplaySelect:
		*d = VariableDisplay(raw)
	default:
		return fmt.Errorf("unmanaged jumper variable display")
	}
	return nil
}
