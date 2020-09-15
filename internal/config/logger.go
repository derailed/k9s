package config

import (
	"github.com/derailed/k9s/internal/client"
)

const (
	// DefaultLoggerTailCount tracks default log tail size.
	DefaultLoggerTailCount = 100
	// MaxLogThreshold sets the max value for log size.
	MaxLogThreshold = 5000
	// DefaultSinceSeconds tracks default log age.
<<<<<<< HEAD
	DefaultSinceSeconds = 60 // all logs
=======
	DefaultSinceSeconds = 300
>>>>>>> 8f7640a75993a24409d6b3d9dd2204c24b7ac1a1
)

// Logger tracks logger options
type Logger struct {
	TailCount      int64 `yaml:"tail"`
	BufferSize     int   `yaml:"buffer"`
	SinceSeconds   int64 `yaml:"sinceSeconds"`
	FullScreenLogs bool  `yaml:"fullScreenLogs"`
	TextWrap       bool  `yaml:"textWrap"`
	ShowTime       bool  `yaml:"showTime"`
}

// NewLogger returns a new instance.
func NewLogger() *Logger {
	return &Logger{
		TailCount:    DefaultLoggerTailCount,
		BufferSize:   MaxLogThreshold,
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
	if l.BufferSize <= 0 || l.BufferSize > MaxLogThreshold {
		l.BufferSize = MaxLogThreshold
	}
	if l.SinceSeconds == 0 {
		l.SinceSeconds = DefaultSinceSeconds
	}
}
