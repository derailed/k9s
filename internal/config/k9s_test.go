package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestK9sValidate(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	m.When(ksMock.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(ksMock.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(ksMock.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)

	c := config.NewK9s()
	c.Validate(ksMock)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 200, c.LogBufferSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sValidateBlank(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	m.When(ksMock.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(ksMock.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(ksMock.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)

	var c config.K9s
	c.Validate(ksMock)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 200, c.LogBufferSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sActiveClusterZero(t *testing.T) {
	setup(t)

	c := config.NewK9s()
	c.CurrentCluster = "fred"
	cl := c.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "default", cl.Namespace.Active)
	assert.Equal(t, 1, len(cl.Namespace.Favorites))
}

func TestK9sActiveClusterBlank(t *testing.T) {
	setup(t)

	var c config.K9s
	cl := c.ActiveCluster()
	assert.Nil(t, cl)
}

func TestK9sActiveCluster(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)
	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))

	cl := cfg.K9s.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "kube-system", cl.Namespace.Active)
	assert.Equal(t, 5, len(cl.Namespace.Favorites))
}
