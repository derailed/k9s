// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

const (
	// SeverityLow tracks low severity.
	SeverityLow SeverityLevel = iota

	// SeverityMedium tracks medium severity level.
	SeverityMedium

	// SeverityHigh tracks high severity level.
	SeverityHigh
)

// SeverityLevel tracks severity levels.
type SeverityLevel int

// Severity tracks a resource severity levels.
type Severity struct {
	Critical int `yaml:"critical"`
	Warn     int `yaml:"warn"`
}

// NewSeverity returns a new instance.
func NewSeverity() *Severity {
	return &Severity{
		Critical: 90,
		Warn:     70,
	}
}

// Validate checks all thresholds and make sure we're cool. If not use defaults.
func (s *Severity) Validate() {
	norm := NewSeverity()
	if !validateRange(s.Warn) {
		s.Warn = norm.Warn
	}
	if !validateRange(s.Critical) {
		s.Critical = norm.Critical
	}
}

func validateRange(v int) bool {
	if v <= 0 || v > 100 {
		return false
	}
	return true
}

// Threshold tracks threshold to alert user when exceeded.
type Threshold map[string]*Severity

// NewThreshold returns a new threshold.
func NewThreshold() Threshold {
	return Threshold{
		"cpu":    NewSeverity(),
		"memory": NewSeverity(),
	}
}

// Validate a namespace is setup correctly.
func (t Threshold) Validate() Threshold {
	for _, k := range []string{"cpu", "memory"} {
		v, ok := t[k]
		if !ok {
			t[k] = NewSeverity()
		} else {
			v.Validate()
		}
	}

	return t
}

// LevelFor returns a defcon level for the current state.
func (t Threshold) LevelFor(k string, v int) SeverityLevel {
	s, ok := t[k]
	if !ok || v < 0 || v > 100 {
		return SeverityLow
	}
	if v >= s.Critical {
		return SeverityHigh
	}
	if v >= s.Warn {
		return SeverityMedium
	}

	return SeverityLow
}

// SeverityColor returns a defcon level associated level.
func (t *Threshold) SeverityColor(k string, v int) string {
	// nolint:exhaustive
	switch t.LevelFor(k, v) {
	case SeverityHigh:
		return "red"
	case SeverityMedium:
		return "orangered"
	default:
		return "green"
	}
}
