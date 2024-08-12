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

	assert.Equal(t, int64(100), l.TailCount)
	assert.Equal(t, 5000, l.BufferSize)
}

func TestLoggerValidate(t *testing.T) {
	var l config.Logger
	l = l.Validate()

	assert.Equal(t, int64(100), l.TailCount)
	assert.Equal(t, 5000, l.BufferSize)
}
