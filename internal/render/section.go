// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

// Level tracks lint check level.
type Level int

const (
	// OkLevel denotes no linting issues.
	OkLevel Level = iota
	// InfoLevel denotes FIY linting issues.
	InfoLevel
	// WarnLevel denotes a warning issue.
	WarnLevel
	// ErrorLevel denotes a serious issue.
	ErrorLevel
)

type (
	// Sections represents a collection of sections.
	Sections []Section

	// Section represents a sanitizer pass.
	Section struct {
		Title   string  `json:"sanitizer" yaml:"sanitizer"`
		GVR     string  `yaml:"gvr" json:"gvr"`
		Outcome Outcome `json:"issues,omitempty" yaml:"issues,omitempty"`
	}

	// Outcome represents a classification of reports outcome.
	Outcome map[string]Issues

	// Issues represents a collection of issues.
	Issues []Issue

	// Issue represents a sanitization issue.
	Issue struct {
		Group   string `yaml:"group" json:"group"`
		GVR     string `yaml:"gvr" json:"gvr"`
		Level   Level  `yaml:"level" json:"level"`
		Message string `yaml:"message" json:"message"`
	}
)
