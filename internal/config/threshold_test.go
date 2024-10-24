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

func TestLevelFor(t *testing.T) {
	uu := map[string]struct {
		k string
		v int
		e config.SeverityLevel
	}{
		"normal": {
			k: "cpu",
			v: 0,
			e: config.SeverityLow,
		},
		"4": {
			k: "cpu",
			v: 71,
			e: config.SeverityMedium,
		},
		"3": {
			k: "cpu",
			v: 75,
			e: config.SeverityMedium,
		},
		"2": {
			k: "cpu",
			v: 80,
			e: config.SeverityMedium,
		},
		"1": {
			k: "cpu",
			v: 100,
			e: config.SeverityHigh,
		},
		"over": {
			k: "cpu",
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
