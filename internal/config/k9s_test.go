// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"errors"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestK9sReload(t *testing.T) {
	config.AppConfigDir = "/tmp/k9s-test"

	cl, ct := "cl-1", "ct-1-1"

	uu := map[string]struct {
		k      *config.K9s
		cl, ct string
		err    error
	}{
		"no-context": {
			k: config.NewK9s(
				mock.NewMockConnection(),
				mock.NewMockKubeSettings(&genericclioptions.ConfigFlags{
					ClusterName: &cl,
					Context:     &ct,
				}),
			),
			err: errors.New(`no context found for: ""`),
		},
		"set-context": {
			k: config.NewK9s(
				mock.NewMockConnection(),
				mock.NewMockKubeSettings(&genericclioptions.ConfigFlags{
					ClusterName: &cl,
					Context:     &ct,
				}),
			),
			ct: "ct-1-1",
			cl: "cl-1",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			_, _ = u.k.ActivateContext(u.ct)
			assert.Equal(t, u.err, u.k.Reload())
			ct, err := u.k.ActiveContext()
			assert.Equal(t, u.err, err)
			if err == nil {
				assert.Equal(t, u.cl, ct.ClusterName)
			}
		})
	}
}

func TestK9sMerge(t *testing.T) {
	cl, ct := "cl-1", "ct-1-1"

	uu := map[string]struct {
		k1, k2 *config.K9s
		ek     *config.K9s
	}{
		"no-opt": {
			k1: config.NewK9s(
				mock.NewMockConnection(),
				mock.NewMockKubeSettings(&genericclioptions.ConfigFlags{
					ClusterName: &cl,
					Context:     &ct,
				}),
			),
			ek: config.NewK9s(
				mock.NewMockConnection(),
				mock.NewMockKubeSettings(&genericclioptions.ConfigFlags{
					ClusterName: &cl,
					Context:     &ct,
				}),
			),
		},
		"override": {
			k1: &config.K9s{
				LiveViewAutoRefresh: false,
				ScreenDumpDir:       "",
				RefreshRate:         0,
				MaxConnRetry:        0,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI:                  config.UI{},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
				ShellPod:            config.ShellPod{},
				ImageScans:          config.ImageScans{},
				Logger:              config.Logger{},
				Thresholds:          nil,
			},
			k2: &config.K9s{
				LiveViewAutoRefresh: true,
				MaxConnRetry:        100,
				ShellPod:            config.NewShellPod(),
			},
			ek: &config.K9s{
				LiveViewAutoRefresh: true,
				ScreenDumpDir:       "",
				RefreshRate:         0,
				MaxConnRetry:        100,
				ReadOnly:            false,
				NoExitOnCtrlC:       false,
				UI:                  config.UI{},
				SkipLatestRevCheck:  false,
				DisablePodCounting:  false,
				ShellPod:            config.NewShellPod(),
				ImageScans:          config.ImageScans{},
				Logger:              config.Logger{},
				Thresholds:          nil,
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			u.k1.Merge(u.k2)
			assert.Equal(t, u.ek, u.k1)
		})
	}
}

func TestContextScreenDumpDir(t *testing.T) {
	cfg := mock.NewMockConfig()
	_, err := cfg.K9s.ActivateContext("ct-1-1")

	assert.NoError(t, err)
	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))
	assert.Equal(t, "/tmp/k9s-test/screen-dumps/cl-1/ct-1-1", cfg.K9s.ContextScreenDumpDir())
}

func TestAppScreenDumpDir(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/configs/k9s.yaml", true))
	assert.Equal(t, "/tmp/k9s-test/screen-dumps", cfg.K9s.AppScreenDumpDir())
}
