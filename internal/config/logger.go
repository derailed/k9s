package config

import (
	"github.com/derailed/k9s/internal/client"
)

const (
	// DefaultLoggerTailCount tracks log tail size.
	DefaultLoggerTailCount = 50
	// DefaultLoggerBufferSize tracks the buffer size.
	DefaultLoggerBufferSize = 1_000
	// MaxLogThreshold sets the max value for log size.
	MaxLogThreshold = 5_000
)

// Logger tracks logger options
type Logger struct {
	TailCount  int `yaml:"tail"`
	BufferSize int `yaml:"buffer"`
}

// NewLogger returns a new instance.
func NewLogger() *Logger {
	return &Logger{
		TailCount:  DefaultLoggerTailCount,
		BufferSize: DefaultLoggerBufferSize,
	}
}

// Validate checks thresholds and make sure we're cool. If not use defaults.
func (l *Logger) Validate(_ client.Connection, _ KubeSettings) {
	if l.TailCount <= 0 {
		l.TailCount = DefaultLoggerTailCount
	}
	if l.TailCount > MaxLogThreshold {
		l.TailCount = MaxLogThreshold
	}
	if l.BufferSize <= 0 {
		l.BufferSize = DefaultLoggerBufferSize
	}
	if l.BufferSize > MaxLogThreshold {
		l.BufferSize = MaxLogThreshold
	}
}
