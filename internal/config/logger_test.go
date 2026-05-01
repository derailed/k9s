// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	l := config.NewLogger()
	l = l.Validate()

	assert.Equal(t, int64(config.DefaultLoggerTailCount), l.TailCount)
	assert.Equal(t, config.MaxLogThreshold, l.BufferSize)
}

func TestLoggerValidate(t *testing.T) {
	var l config.Logger
	l = l.Validate()

	assert.Equal(t, int64(config.DefaultLoggerTailCount), l.TailCount)
	assert.Equal(t, config.MaxLogThreshold, l.BufferSize)
}

func TestLoggerValidateOverMax(t *testing.T) {
	l := config.Logger{
		TailCount:  20_000,
		BufferSize: 20_000,
	}
	l = l.Validate()

	assert.Equal(t, int64(config.MaxLogThreshold), l.TailCount)
	assert.Equal(t, config.MaxLogThreshold, l.BufferSize)
}

func TestLoggerValidateCustom(t *testing.T) {
	l := config.Logger{
		TailCount:  500,
		BufferSize: 8000,
	}
	l = l.Validate()

	assert.Equal(t, int64(500), l.TailCount)
	assert.Equal(t, 8000, l.BufferSize)
}
