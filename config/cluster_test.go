package config_test

import (
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/derailed/k9s/config"
	"github.com/stretchr/testify/assert"
)

func TestClusterValidate(t *testing.T) {
	setup(t)

	ciMock := NewMockClusterInfo()
	m.When(ciMock.AllNamespacesOrDie()).ThenReturn([]string{"ns1", "ns2", "default"})

	c := config.NewCluster()
	c.Validate(ciMock)

	assert.Equal(t, "po", c.View.Active)
	assert.Equal(t, "default", c.Namespace.Active)
	assert.Equal(t, 2, len(c.Namespace.Favorites))
	assert.Equal(t, []string{"all", "default"}, c.Namespace.Favorites)
}
