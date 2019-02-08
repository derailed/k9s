package config_test

import (
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/derailed/k9s/config"
	"github.com/stretchr/testify/assert"
)

func TestK9sValidate(t *testing.T) {
	setup(t)
	ci := NewMockClusterInfo()
	m.When(ci.AllClustersOrDie()).ThenReturn([]string{"c1", "c2"})
	m.When(ci.ActiveClusterOrDie()).ThenReturn("c1")

	c := config.NewK9s()
	c.Validate(ci)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 200, c.LogBufferSize)
	assert.Equal(t, "c1", c.Context.Active)
	assert.Equal(t, 1, len(c.Context.Clusters))
}
