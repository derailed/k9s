// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

const (
	// DefaultLoggerTailCount tracks default log tail size.
	DefaultLoggerTailCount = 100

	// MaxLogThreshold sets the max value for log size.
	MaxLogThreshold = 5000

	// DefaultSinceSeconds tracks default log age.
	DefaultSinceSeconds = -1 // tail logs by default
)

// Logger tracks logger options.
type Logger struct {
	TailCount    int64 `json:"tail" yaml:"tail"`
	BufferSize   int   `json:"buffer" yaml:"buffer"`
	SinceSeconds int64 `json:"sinceSeconds" yaml:"sinceSeconds"`
	TextWrap     bool  `json:"textWrap" yaml:"textWrap"`
	ShowTime     bool  `json:"showTime" yaml:"showTime"`
}

// NewLogger returns a new instance.
func NewLogger() Logger {
	return Logger{
		TailCount:    DefaultLoggerTailCount,
		BufferSize:   MaxLogThreshold,
		SinceSeconds: DefaultSinceSeconds,
	}
}

// Validate checks thresholds and make sure we're cool. If not use defaults.
func (l Logger) Validate() Logger {
	if l.TailCount <= 0 {
		l.TailCount = DefaultLoggerTailCount
	}
	if l.TailCount > MaxLogThreshold {
		l.TailCount = MaxLogThreshold
	}
	if l.BufferSize <= 0 || l.BufferSize > MaxLogThreshold {
		l.BufferSize = MaxLogThreshold
	}
	if l.SinceSeconds == 0 {
		l.SinceSeconds = DefaultSinceSeconds
	}

	return l
}
