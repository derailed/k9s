// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSeverityValidate(t *testing.T) {
	uu := map[string]struct {
		d, e *config.Severity
	}{
		"default": {
			d: config.NewSeverity(),
			e: config.NewSeverity(),
		},
		"toast": {
			d: &config.Severity{Warn: 10},
			e: &config.Severity{Warn: 10, Critical: 90},
		},
		"negative": {
			d: &config.Severity{Warn: -1},
			e: config.NewSeverity(),
		},
		"out-of-range": {
			d: &config.Severity{Warn: 150},
			e: config.NewSeverity(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.d.Validate()
			assert.Equal(t, u.e, u.d)
		})
	}
}

func TestThresholdValidate(t *testing.T) {
	custom := func() *config.Severity {
		return &config.Severity{Warn: 10, Critical: 20}
	}

	uu := map[string]struct {
		d, e config.Threshold
	}{
		"nil-cpu": {
			d: config.Threshold{config.CPU: nil, config.MEM: custom()},
			e: config.Threshold{config.CPU: config.NewSeverity(), config.MEM: custom()},
		},
		"nil-memory": {
			d: config.Threshold{config.CPU: custom(), config.MEM: nil},
			e: config.Threshold{config.CPU: custom(), config.MEM: config.NewSeverity()},
		},
		"nil-both": {
			d: config.Threshold{config.CPU: nil, config.MEM: nil},
			e: config.NewThreshold(),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, u.d.Validate())
		})
	}
}

func TestLevelFor(t *testing.T) {
	uu := map[string]struct {
		k string
		v int
		e config.SeverityLevel
	}{
		"normal": {
			k: config.CPU,
			v: 0,
			e: config.SeverityLow,
		},
		"4": {
			k: config.CPU,
			v: 71,
			e: config.SeverityMedium,
		},
		"3": {
			k: config.CPU,
			v: 75,
			e: config.SeverityMedium,
		},
		"2": {
			k: config.CPU,
			v: 80,
			e: config.SeverityMedium,
		},
		"1": {
			k: config.CPU,
			v: 100,
			e: config.SeverityHigh,
		},
		"over": {
			k: config.CPU,
			v: 150,
			e: config.SeverityLow,
		},
		"over-mem": {
			k: config.MEM,
			v: 150,
			e: config.SeverityLow,
		},
	}

	o := config.NewThreshold()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, o.LevelFor(u.k, u.v))
		})
	}
}

func TestLevelForNilSeverity(t *testing.T) {
	o := config.Threshold{config.CPU: nil, config.MEM: nil}
	for _, k := range []string{config.CPU, config.MEM} {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, config.SeverityLow, o.LevelFor(k, 95))
		})
	}
}
