package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/client"
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
	cfgFile, ctx, cluster, ns := "testdata/kubeconfig-test.yml", "test2", "cluster2", "ns2"
	uu := map[string]struct {
		flags                       *genericclioptions.ConfigFlags
		issue                       bool
		context, cluster, namespace string
	}{
		"plain": {
			flags:     &genericclioptions.ConfigFlags{KubeConfig: &cfgFile},
			issue:     false,
			context:   "test1",
			cluster:   "cluster1",
			namespace: "ns1",
		},
		"overrideNS": {
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

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			mc := NewMockConnection()
			m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)
			mk := newMockSettings(u.flags)
			cfg := config.NewConfig(mk)

			err := cfg.Refine(u.flags, nil, client.NewConfig(u.flags))
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
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.Validate()
}

func TestConfigLoad(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))

	assert.Equal(t, 2, cfg.K9s.RefreshRate)
	assert.Equal(t, 2000, cfg.K9s.Logger.BufferSize)
	assert.Equal(t, int64(200), cfg.K9s.Logger.TailCount)
	assert.Equal(t, "minikube", cfg.K9s.CurrentContext)
	assert.Equal(t, "minikube", cfg.K9s.CurrentCluster)
	assert.NotNil(t, cfg.K9s.Clusters)
	assert.Equal(t, 2, len(cfg.K9s.Clusters))

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

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	assert.NotNil(t, cfg.CurrentCluster())
	assert.Equal(t, "kube-system", cfg.CurrentCluster().Namespace.Active)
	assert.Equal(t, "ctx", cfg.CurrentCluster().View.Active)
}

func TestConfigActiveNamespace(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	assert.Equal(t, "kube-system", cfg.ActiveNamespace())
}

func TestConfigActiveNamespaceBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigSetActiveNamespace(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	assert.Nil(t, cfg.SetActiveNamespace("default"))
	assert.Equal(t, "default", cfg.ActiveNamespace())
}

func TestConfigActiveView(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	assert.Equal(t, "ctx", cfg.ActiveView())
}

func TestConfigActiveViewBlank(t *testing.T) {
	cfg := config.Config{K9s: new(config.K9s)}
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigSetActiveView(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.SetActiveView("po")
	assert.Equal(t, "po", cfg.ActiveView())
}

func TestConfigFavNamespaces(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)

	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	expectedNS := []string{"default", "kube-public", "istio-system", "all", "kube-system"}
	assert.Equal(t, expectedNS, cfg.FavNamespaces())
}

func TestConfigLoadOldCfg(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.Nil(t, cfg.Load("testdata/k9s_old.yml"))
}

func TestConfigLoadCrap(t *testing.T) {
	mk := NewMockKubeSettings()
	cfg := config.NewConfig(mk)
	assert.NotNil(t, cfg.Load("testdata/k9s_not_there.yml"))
}

func TestConfigSaveFile(t *testing.T) {
	mc := NewMockConnection()
	m.When(mc.ValidNamespaces()).ThenReturn(namespaces(), nil)

	mk := NewMockKubeSettings()
	m.When(mk.CurrentContextName()).ThenReturn("minikube", nil)
	m.When(mk.CurrentClusterName()).ThenReturn("minikube", nil)
	m.When(mk.CurrentNamespaceName()).ThenReturn("default", nil)
	m.When(mk.ClusterNames()).ThenReturn(map[string]struct{}{"minikube": {}, "fred": {}, "blee": {}}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	cfg := config.NewConfig(mk)
	cfg.SetConnection(mc)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.K9s.RefreshRate = 100
	cfg.K9s.ReadOnly = true
	cfg.K9s.Logger.TailCount = 500
	cfg.K9s.Logger.BufferSize = 800
	cfg.K9s.CurrentContext = "blee"
	cfg.K9s.CurrentCluster = "blee"
	cfg.Validate()
	path := filepath.Join("/tmp", "k9s.yml")
	err := cfg.SaveFile(path)
	assert.Nil(t, err)

	raw, err := os.ReadFile(path)
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
	m.When(mk.ClusterNames()).ThenReturn(map[string]struct{}{"blee": {}}, nil)
	m.When(mk.NamespaceNames(namespaces())).ThenReturn([]string{"default"})

	cfg := config.NewConfig(mk)
	cfg.SetConnection(mc)
	assert.Nil(t, cfg.Load("testdata/k9s.yml"))
	cfg.Reset()
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yml")
	err := cfg.SaveFile(path)
	assert.Nil(t, err)

	raw, err := os.ReadFile(path)
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

type mockSettings struct {
	flags *genericclioptions.ConfigFlags
}

var _ config.KubeSettings = (*mockSettings)(nil)

func newMockSettings(flags *genericclioptions.ConfigFlags) *mockSettings {
	return &mockSettings{flags: flags}
}
func (m *mockSettings) CurrentContextName() (string, error) {
	return *m.flags.Context, nil
}
func (m *mockSettings) CurrentClusterName() (string, error) { return "", nil }
func (m *mockSettings) CurrentNamespaceName() (string, error) {
	return *m.flags.Namespace, nil
}
func (m *mockSettings) ClusterNames() (map[string]struct{}, error) { return nil, nil }

// ----------------------------------------------------------------------------
// Test Data...

var expectedConfig = `k9s:
  liveViewAutoRefresh: true
  refreshRate: 100
  maxConnRetry: 5
  enableMouse: false
  headless: false
  logoless: false
  crumbsless: false
  readOnly: true
  noExitOnCtrlC: false
  noIcons: false
  skipLatestRevCheck: false
  logger:
    tail: 500
    buffer: 800
    sinceSeconds: 300
    fullScreenLogs: false
    textWrap: false
    showTime: false
  currentContext: blee
  currentCluster: blee
  clusters:
    blee:
      namespace:
        active: default
        lockFavorites: false
        favorites:
        - default
      view:
        active: po
      featureGates:
        nodeShell: false
      shellPod:
        image: busybox:1.35.0
        command: []
        args: []
        namespace: default
        limits:
          cpu: 100m
          memory: 100Mi
        labels: {}
      portForwardAddress: localhost
    fred:
      namespace:
        active: default
        lockFavorites: false
        favorites:
        - default
        - kube-public
        - istio-system
        - all
        - kube-system
      view:
        active: po
      featureGates:
        nodeShell: false
      shellPod:
        image: busybox:1.35.0
        command: []
        args: []
        namespace: default
        limits:
          cpu: 100m
          memory: 100Mi
        labels: {}
      portForwardAddress: localhost
    minikube:
      namespace:
        active: kube-system
        lockFavorites: false
        favorites:
        - default
        - kube-public
        - istio-system
        - all
        - kube-system
      view:
        active: ctx
      featureGates:
        nodeShell: false
      shellPod:
        image: busybox:1.35.0
        command: []
        args: []
        namespace: default
        limits:
          cpu: 100m
          memory: 100Mi
        labels: {}
      portForwardAddress: localhost
  thresholds:
    cpu:
      critical: 90
      warn: 70
    memory:
      critical: 90
      warn: 70
  screenDumpDir: /tmp
`

var resetConfig = `k9s:
  liveViewAutoRefresh: true
  refreshRate: 2
  maxConnRetry: 5
  enableMouse: false
  headless: false
  logoless: false
  crumbsless: false
  readOnly: false
  noExitOnCtrlC: false
  noIcons: false
  skipLatestRevCheck: false
  logger:
    tail: 200
    buffer: 2000
    sinceSeconds: 300
    fullScreenLogs: false
    textWrap: false
    showTime: false
  currentContext: blee
  currentCluster: blee
  clusters:
    blee:
      namespace:
        active: default
        lockFavorites: false
        favorites:
        - default
      view:
        active: po
      featureGates:
        nodeShell: false
      shellPod:
        image: busybox:1.35.0
        command: []
        args: []
        namespace: default
        limits:
          cpu: 100m
          memory: 100Mi
        labels: {}
      portForwardAddress: localhost
  thresholds:
    cpu:
      critical: 90
      warn: 70
    memory:
      critical: 90
      warn: 70
  screenDumpDir: /tmp
`
