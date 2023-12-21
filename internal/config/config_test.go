// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	m "github.com/petergtz/pegomock"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func init() {
	zerolog.SetGlobalLevel(zerolog.FatalLevel)
}

func TestConfigRefine(t *testing.T) {
	var (
		cfgFile          = "testdata/kubeconfig-test.yaml"
		ctx, cluster, ns = "ct-1-1", "cl-1", "ns-1"
	)

	uu := map[string]struct {
		flags                       *genericclioptions.ConfigFlags
		issue                       bool
		context, cluster, namespace string
	}{
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
			cfg := mock.NewMockConfig()

			err := cfg.Refine(u.flags, nil, client.NewConfig(u.flags))
			if u.issue {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, u.context, cfg.K9s.ActiveContextName())
				assert.Equal(t, u.namespace, cfg.ActiveNamespace())
			}
		})
	}
}

func TestConfigValidate(t *testing.T) {
	cfg := mock.NewMockConfig()
	cfg.SetConnection(mock.NewMockConnection())

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	cfg.Validate()
}

func TestConfigLoad(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	assert.Equal(t, 2, cfg.K9s.RefreshRate)
	assert.Equal(t, 2000, cfg.K9s.Logger.BufferSize)
	assert.Equal(t, int64(200), cfg.K9s.Logger.TailCount)
}

func TestConfigLoadOldCfg(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s_old.yaml"))
}

func TestConfigLoadCrap(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.NotNil(t, cfg.Load("testdata/k9s_not_there.yaml"))
}

func TestConfigSaveFile(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))

	cfg.K9s.RefreshRate = 100
	cfg.K9s.ReadOnly = true
	cfg.K9s.Logger.TailCount = 500
	cfg.K9s.Logger.BufferSize = 800
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yaml")
	err := cfg.SaveFile(path)
	assert.Nil(t, err)
	raw, err := os.ReadFile(path)
	assert.Nil(t, err)
	assert.Equal(t, expectedConfig, string(raw))
}

func TestConfigReset(t *testing.T) {
	cfg := mock.NewMockConfig()
	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	cfg.Reset()
	cfg.Validate()

	path := filepath.Join("/tmp", "k9s.yaml")
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

// ----------------------------------------------------------------------------
// Test Data...

var expectedConfig = `k9s:
  liveViewAutoRefresh: true
  screenDumpDir: /tmp
  refreshRate: 100
  maxConnRetry: 5
  readOnly: true
  noExitOnCtrlC: false
  ui:
    enableMouse: false
    headless: false
    logoless: false
    crumbsless: false
    noIcons: false
  skipLatestRevCheck: false
  disablePodCounting: false
  shellPod:
    image: busybox:1.35.0
    namespace: default
    limits:
      cpu: 100m
      memory: 100Mi
  imageScans:
    enable: false
    blackList:
      namespaces: []
      labels: {}
  logger:
    tail: 500
    buffer: 800
    sinceSeconds: -1
    fullScreenLogs: false
    textWrap: false
    showTime: false
  thresholds:
    cpu:
      critical: 90
      warn: 70
    memory:
      critical: 90
      warn: 70
`

var resetConfig = `k9s:
  liveViewAutoRefresh: true
  screenDumpDir: /tmp
  refreshRate: 2
  maxConnRetry: 5
  readOnly: false
  noExitOnCtrlC: false
  ui:
    enableMouse: false
    headless: false
    logoless: false
    crumbsless: false
    noIcons: false
  skipLatestRevCheck: false
  disablePodCounting: false
  shellPod:
    image: busybox:1.35.0
    namespace: default
    limits:
      cpu: 100m
      memory: 100Mi
  imageScans:
    enable: false
    blackList:
      namespaces: []
      labels: {}
  logger:
    tail: 200
    buffer: 2000
    sinceSeconds: -1
    fullScreenLogs: false
    textWrap: false
    showTime: false
  thresholds:
    cpu:
      critical: 90
      warn: 70
    memory:
      critical: 90
      warn: 70
`
