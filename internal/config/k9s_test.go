package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestK9sValidate(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	c := config.NewK9s()
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 50, c.Logger.TailCount)
	assert.Equal(t, 1_000, c.Logger.BufferSize)
	assert.Equal(t, "ctx1", c.CurrentContext)
	assert.Equal(t, "c1", c.CurrentCluster)
	assert.Equal(t, 1, len(c.Clusters))
	_, ok := c.Clusters[c.CurrentCluster]
	assert.True(t, ok)
}

func TestK9sValidateBlank(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("ctx1", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("c1", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	var c config.K9s
	c.Validate(mc, mk)

	assert.Equal(t, 2, c.RefreshRate)
	assert.Equal(t, 50, c.Logger.TailCount)
	assert.Equal(t, 1_000, c.Logger.BufferSize)
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
	assert.Equal(t, config.NewCluster(), cl)
}

func TestK9sActiveCluster(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))

	cl := cfg.K9s.ActiveCluster()
	assert.NotNil(t, cl)
	assert.Equal(t, "kube-system", cl.Namespace.Active)
	assert.Equal(t, 5, len(cl.Namespace.Favorites))
}
