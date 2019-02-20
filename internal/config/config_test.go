package config_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/stretchr/testify/assert"
)

func TestConfigValidate(t *testing.T) {
	setup(t)

	ksMock := NewMockKubeSettings()
	m.When(ksMock.NamespaceNames()).ThenReturn([]string{"ns1", "ns2", "default"}, nil)
	m.When(ksMock.ClusterNames()).ThenReturn([]string{"c1", "c2"}, nil)

	cfg := config.NewConfig(ksMock)
	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.Validate()
}

func TestConfigLoad(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)
	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))

	assert.Equal(t, 2, cfg.K9s.RefreshRate)
	assert.Equal(t, 200, cfg.K9s.LogBufferSize)
	assert.Equal(t, "minikube", cfg.K9s.CurrentContext)
	assert.Equal(t, "minikube", cfg.K9s.CurrentCluster)
	assert.NotNil(t, cfg.K9s.Clusters)
	assert.Equal(t, 2, len(cfg.K9s.Clusters))
	assert.Equal(t, 0, len(cfg.K9s.Aliases))

	nn := []string{
		"default",
		"kube-public",
		"istio-system",
		"all",
		"kube-system",
	}

	assert.Equal(t, "kube-system", cfg.K9s.Clusters["minikube"].Namespace.Active)
	assert.Equal(t, nn, cfg.K9s.Clusters["minikube"].Namespace.Favorites)
	assert.Equal(t, "ctx", cfg.K9s.Clusters["minikube"].View.Active)
}

func TestConfigCurrentCluster(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.NotNil(t, cfg.CurrentCluster())
	assert.Equal(t, "kube-system", cfg.CurrentCluster().Namespace.Active)
	assert.Equal(t, "ctx", cfg.CurrentCluster().View.Active)
}

func TestConfigActiveNamespace(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.Equal(t, "kube-system", cfg.ActiveNamespace())
}

func TestConfigActiveNamespaceBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigSetActiveNamespace(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.SetActiveNamespace("default")
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigActiveView(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.Equal(t, "ctx", cfg.ActiveView())
}

func TestConfigActiveViewBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigSetActiveView(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.SetActiveView("po")
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigFavNamespaces(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	expectedNS := []string{"default", "kube-public", "istio-system", "all", "kube-system"}
	assert.Equal(t, expectedNS, cfg.FavNamespaces())
}

func TestConfigLoadOldCfg(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)
	assert.Nil(t, cfg.Load("test_assets/k9s_old.yml"))
}

func TestConfigLoadCrap(t *testing.T) {
	ksMock := NewMockKubeSettings()
	cfg := config.NewConfig(ksMock)
	assert.NotNil(t, cfg.Load("test_assets/k9s_not_there.yml"))
}

func TestConfigSaveFile(t *testing.T) {
	ksMock := NewMockKubeSettings()
	m.When(ksMock.CurrentContextName()).ThenReturn("minikube", nil)
	m.When(ksMock.CurrentClusterName()).ThenReturn("minikube", nil)
	m.When(ksMock.CurrentNamespaceName()).ThenReturn("default", nil)
	m.When(ksMock.ClusterNames()).ThenReturn([]string{"minikube", "fred", "blee"}, nil)

	cfg := config.NewConfig(ksMock)
	cfg.Load("test_assets/k9s.yml")
	cfg.K9s.RefreshRate = 100
	cfg.K9s.LogBufferSize = 500
	cfg.K9s.CurrentContext = "blee"
	cfg.K9s.CurrentCluster = "blee"
	cfg.Validate()
	path := filepath.Join("/tmp", "k9s.yml")
	err := cfg.SaveFile(path)
	assert.Nil(t, err)

	raw, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, string(raw))
}

func TestConfigReset(t *testing.T) {
	ksMock := NewMockKubeSettings()
	m.When(ksMock.CurrentContextName()).ThenReturn("blee", nil)
	m.When(ksMock.CurrentClusterName()).ThenReturn("blee", nil)
	m.When(ksMock.CurrentNamespaceName()).ThenReturn("default", nil)
	m.When(ksMock.ClusterNames()).ThenReturn([]string{"blee"}, nil)

	cfg := config.NewConfig(ksMock)
	cfg.Load("test_assets/k9s.yml")
	cfg.Reset()
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yml")
	err := cfg.SaveFile(path)
	assert.Nil(t, err)

	raw, err := ioutil.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, resetConfig, string(raw))
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
  currentContext: blee
  currentCluster: blee
  clusters:
    blee:
      namespace:
        active: default
        favorites:
        - default
      view:
        active: po
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

var resetConfig = `k9s:
  refreshRate: 2
  logBufferSize: 200
  currentContext: blee
  currentCluster: blee
  clusters:
    blee:
      namespace:
        active: default
        favorites:
        - default
      view:
        active: po
`
