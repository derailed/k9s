// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

// StatusIndicator tracks custom status indicator config
type StatusIndicatorConf struct {
	// Format Represents the formatting template
	Format string `json:"format" yaml:"format"`

	// Fields contains the fields in an ordered manner that should appear in the status indicator
	Fields []string `json:"fields" yaml:"fields"`
}

func NewStatusIndicatorConf() StatusIndicatorConf {
	return StatusIndicatorConf{
		Format: "[orange::b]K9s [aqua::]%s [white::]%s:%s:%s [lawngreen::]%s[white::]::[darkturquoise::]%s",
		Fields: []string{"K9SVER", "CONTEXT", "CLUSTER", "K8SVER", "CPU", "MEMORY"},
	}
}
