package views

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestNewApp(t *testing.T) {
	a := NewApp(config.NewConfig(ks{}))
	a.Init("blee", 10)

	assert.Equal(t, 10, len(a.actions))
	assert.Equal(t, false, a.hasSkins)
}
