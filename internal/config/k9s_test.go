package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestK9sValidate(t *testing.T) {
	mc := NewMockConnection()
	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)

	c := config.NewK9s()
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 1000, c.LogBufferSize)
	assert.Equal(t, 200, c.LogRequestSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sValidateBlank(t *testing.T) {
	mc := NewMockConnection()
	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)

	var c config.K9s
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 1000, c.LogBufferSize)
	assert.Equal(t, 200, c.LogRequestSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sActiveClusterZero(t *testing.T) {
	c := config.NewK9s()
	c.CurrentCluster = "fred"
	cl := c.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "default", cl.Namespace.Active)
	assert.Equal(t, 1, len(cl.Namespace.Favorites))
}

func TestK9sActiveClusterBlank(t *testing.T) {
	var c config.K9s
	cl := c.ActiveCluster()
	assert.Nil(t, cl)
}

func TestK9sActiveCluster(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))

	cl := cfg.K9s.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "kube-system", cl.Namespace.Active)
	assert.Equal(t, 5, len(cl.Namespace.Favorites))
}
