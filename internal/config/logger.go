package config

import (
	"github.com/derailed/k9s/internal/client"
)

const (
	// DefaultLoggerTailCount tracks default log tail size.
	DefaultLoggerTailCount = 100
	// DefaultLoggerBufferSize tracks default view buffer size.
	DefaultLoggerBufferSize = 1_000
	// MaxLogThreshold sets the max value for log size.
	MaxLogThreshold = 1_000
	// DefaultSinceSeconds tracks default log age.
	DefaultSinceSeconds = 5 * 60 // 5mins
)

// Logger tracks logger options
type Logger struct {
	TailCount    int64 `yaml:"tail"`
	BufferSize   int   `yaml:"buffer"`
	SinceSeconds int64 `yaml:"sinceSeconds"`
}

// NewLogger returns a new instance.
func NewLogger() *Logger {
	return &Logger{
		TailCount:    DefaultLoggerTailCount,
		BufferSize:   DefaultLoggerBufferSize,
		SinceSeconds: DefaultSinceSeconds,
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
	if l.SinceSeconds == 0 {
		l.SinceSeconds = DefaultSinceSeconds
	}
}
