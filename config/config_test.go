package config_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	m "github.com/petergtz/pegomock"
	"github.com/derailed/k9s/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	setup(t)

	assert.Nil(t, config.Load("test_assets/k9s.yml"))

	ciMock := NewMockClusterInfo()
	m.When(ciMock.AllNamespacesOrDie()).ThenReturn([]string{"ns1", "ns2", "default"})
	m.When(ciMock.AllClustersOrDie()).ThenReturn([]string{"c1", "c2"})

	config.Root.Validate(ciMock)

}

func TestConfigLoad(t *testing.T) {
	assert.Nil(t, config.Load("test_assets/k9s.yml"))
	assert.Equal(t, 2, config.Root.K9s.RefreshRate)
	assert.Equal(t, 200, config.Root.K9s.LogBufferSize)

	ctx := config.Root.K9s.Context
	assert.Equal(t, "minikube", ctx.Active)
	assert.NotNil(t, ctx.Clusters)

	nn := []string{
		"default",
		"kube-public",
		"istio-system",
		"all",
		"kube-system",
	}
	assert.Equal(t, "kube-system", ctx.Clusters["minikube"].Namespace.Active)
	assert.Equal(t, nn, ctx.Clusters["minikube"].Namespace.Favorites)
	assert.Equal(t, "ctx", ctx.Clusters["minikube"].View.Active)
}

func TestConfigLoadOldCfg(t *testing.T) {
	assert.Nil(t, config.Load("test_assets/k9s_old.yml"))
}

func TestConfigSaveFile(t *testing.T) {
	config.Load("test_assets/k9s.yml")

	config.Root.K9s.RefreshRate = 100
	config.Root.K9s.LogBufferSize = 500
	config.Root.K9s.Context.Active = "fred"

	path := filepath.Join("/tmp", "k9s.yml")
	err := config.Root.SaveFile(path)
	assert.Nil(t, err)

	raw, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, string(raw))
}

// Helpers...

func setup(t *testing.T) {
	m.RegisterMockTestingT(t)
	m.RegisterMockFailHandler(func(m string, i ...int) {
		fmt.Println("Boom!", m, i)
	})
}

// ----------------------------------------------------------------------------
// Test Data...

var expectedConfig = `k9s:
  refreshRate: 100
  logBufferSize: 500
  context:
    active: fred
    clusters:
      fred:
        namespace:
          active: default
          favorites:
          - default
          - kube-public
          - istio-system
          - all
          - kube-system
        view:
          active: po
      minikube:
        namespace:
          active: kube-system
          favorites:
          - default
          - kube-public
          - istio-system
          - all
          - kube-system
        view:
          active: ctx
`
