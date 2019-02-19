package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestClusterValidate(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2", "default"}, nil)

	c := config.NewCluster()
	c.Validate(ksMock)

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Equal(t, 1, len(c.Namespace.Favorites))
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}

func TestClusterValidateEmpty(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2", "default"}, nil)

	var c config.Cluster
	c.Validate(ksMock)

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Equal(t, 1, len(c.Namespace.Favorites))
	assert.Equal(t, []string{"default"}, c.Namespace.Favorites)
}
