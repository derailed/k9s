package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestPluginLoad(t *testing.T) {
	p := config.NewPlugins()
	assert.Nil(t, p.LoadPlugins("test_assets/plugin.yml"))

	assert.Equal(t, 1, len(p.Plugin))
}
