package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewLogger(t *testing.T) {
	l := config.NewLogger()
	l.Validate(nil, nil)

	assert.Equal(t, 50, l.TailCount)
	assert.Equal(t, 1_000, l.BufferSize)
}

func TestLoggerValidate(t *testing.T) {
	var l config.Logger
	l.Validate(nil, nil)

	assert.Equal(t, 50, l.TailCount)
	assert.Equal(t, 1_000, l.BufferSize)
}
