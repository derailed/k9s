package config_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	m "github.com/petergtz/pegomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestConfigRefine(t *testing.T) {
	cfgFile, ctx, cluster, ns := "test_assets/kubeconfig-test.yml", "test", "c1", "ns1"
	uu := map[string]struct {
		flags                       *genericclioptions.ConfigFlags
		issue                       bool
		context, cluster, namespace string
	}{
		"kubeconfig": {
			flags:     &genericclioptions.ConfigFlags{KubeConfig: &cfgFile},
			issue:     false,
			context:   "test",
			cluster:   "testCluster",
			namespace: "testNS",
		},
		"override": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig:  &cfgFile,
				Context:     &ctx,
				ClusterName: &cluster,
				Namespace:   &ns,
			},
			issue:     false,
			context:   ctx,
			cluster:   cluster,
			namespace: ns,
		},
		"badContext": {
			flags: &genericclioptions.ConfigFlags{
				KubeConfig:  &cfgFile,
				Context:     &ns,
				ClusterName: &cluster,
				Namespace:   &ns,
			},
			issue: true,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			mc := NewMockConnection()
			m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)
			mk := NewMockKubeSettings()
			m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})
			cfg := config.NewConfig(mk)
			err := cfg.Refine(u.flags)

			if u.issue {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, u.context, cfg.K9s.CurrentContext)
				assert.Equal(t, u.cluster, cfg.K9s.CurrentCluster)
				assert.Equal(t, u.namespace, cfg.ActiveNamespace())
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	cfg := config.NewConfig(mk)
	cfg.SetConnection(mc)
	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.Validate()
	// mc.VerifyWasCalledOnce().ValidNamespaces()
}

func TestConfigLoad(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
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
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.NotNil(t, cfg.CurrentCluster())
	assert.Equal(t, "kube-system", cfg.CurrentCluster().Namespace.Active)
	assert.Equal(t, "ctx", cfg.CurrentCluster().View.Active)
}

func TestConfigActiveNamespace(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.Equal(t, "kube-system", cfg.ActiveNamespace())
}

func TestConfigActiveNamespaceBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigSetActiveNamespace(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.SetActiveNamespace("default")
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigActiveView(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	assert.Equal(t, "ctx", cfg.ActiveView())
}

func TestConfigActiveViewBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigSetActiveView(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	cfg.SetActiveView("po")
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigFavNamespaces(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("test_assets/k9s.yml"))
	expectedNS := []string{"default", "kube-public", "istio-system", "all", "kube-system"}
	assert.Equal(t, expectedNS, cfg.FavNamespaces())
}

func TestConfigLoadOldCfg(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("test_assets/k9s_old.yml"))
}

func TestConfigLoadCrap(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.NotNil(t, cfg.Load("test_assets/k9s_not_there.yml"))
}

func TestConfigSaveFile(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("minikube", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("minikube", nil)
	m.When(mk.CurrentNamespaceName()).ThenReturn("default", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"minikube", "fred", "blee"}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	cfg := config.NewConfig(mk)
	cfg.SetConnection(mc)
	cfg.Load("test_assets/k9s.yml")
	cfg.K9s.RefreshRate = 100
	cfg.K9s.LogBufferSize = 500
	cfg.K9s.LogRequestSize = 100
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
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("blee", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("blee", nil)
	m.When(mk.CurrentNamespaceName()).ThenReturn("default", nil)
	m.When(mk.ClusterNames()).ThenReturn([]string{"blee"}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	cfg := config.NewConfig(mk)
	cfg.SetConnection(mc)
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

func TestSetup(t *testing.T) {
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
  logRequestSize: 100
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
  logRequestSize: 200
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
