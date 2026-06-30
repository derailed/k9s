// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	slog.SetDefault(slog.New(slog.DiscardHandler))
}

func TestCustomViewLoad(t *testing.T) {
	uu := map[string]struct {
		cv   *config.CustomView
		path string
		key  string
		e    []string
	}{
		"empty": {},

		"gvr": {
			path: "testdata/views/views.yaml",
			key:  client.PodGVR.String(),
			e:    []string{"NAMESPACE", "NAME", "AGE", "IP"},
		},

		"gvr+ns": {
			path: "testdata/views/views.yaml",
			key:  "v1/pods@default",
			e:    []string{"NAME", "IP", "AGE"},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			cfg := config.NewCustomView()

			require.NoError(t, cfg.Load(u.path))
			assert.Equal(t, u.e, cfg.Views[u.key].Columns)
		})
	}
}

func TestWorkloadGVRsLoad(t *testing.T) {
	cfg := config.NewCustomView()
	require.NoError(t, cfg.Load("testdata/views/workloads.yaml"))

	gvrs, ok := cfg.WorkloadGVRs(config.DefaultWorkloadGVRs)
	require.True(t, ok)
	assert.Equal(t, []string{"apps/v1/deployments", "v1/pods", "examples.demo.io/v1/foos"}, gvrs)

	rks, ok := cfg.WorkloadGVRs("rks")
	require.True(t, ok)
	assert.Equal(t, []string{"rks.io/v1/rksinstances", "rks.io/v1/workerinstances", "v1/pods"}, rks)
}

func TestWorkloadGVRs(t *testing.T) {
	uu := map[string]struct {
		cv   *config.CustomView
		name string
		e    []string
		ok   bool
	}{
		"nil": {
			cv: nil,
		},
		"empty": {
			cv: config.NewCustomView(),
		},
		"default-fallback-on-blank-name": {
			cv: func() *config.CustomView {
				cv := config.NewCustomView()
				cv.Workloads[config.DefaultWorkloadGVRs] = []string{"v1/pods"}
				return cv
			}(),
			name: "",
			e:    []string{"v1/pods"},
			ok:   true,
		},
		"named": {
			cv: func() *config.CustomView {
				cv := config.NewCustomView()
				cv.Workloads["rks"] = []string{"rks.io/v1/rksinstances"}
				return cv
			}(),
			name: "rks",
			e:    []string{"rks.io/v1/rksinstances"},
			ok:   true,
		},
		"unknown-name": {
			cv: func() *config.CustomView {
				cv := config.NewCustomView()
				cv.Workloads[config.DefaultWorkloadGVRs] = []string{"v1/pods"}
				return cv
			}(),
			name: "nope",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			gvrs, ok := u.cv.WorkloadGVRs(u.name)
			assert.Equal(t, u.ok, ok)
			assert.Equal(t, u.e, gvrs)
		})
	}
}

func TestViewSettingEquals(t *testing.T) {
	uu := map[string]struct {
		v1, v2 *config.ViewSetting
		e      bool
	}{
		"v1-nil-v2-nil": {
			e: true,
		},

		"v1-v2-empty": {
			v1: new(config.ViewSetting),
			v2: new(config.ViewSetting),
			e:  true,
		},

		"v1-nil": {
			v1: new(config.ViewSetting),
		},

		"nil-v2": {
			v2: new(config.ViewSetting),
		},

		"v1-v2-blank": {
			v1: &config.ViewSetting{
				Columns: []string{"A"},
			},
			v2: new(config.ViewSetting),
		},

		"v1-v2-nil": {
			v1: &config.ViewSetting{
				Columns: []string{"A"},
			},
		},

		"same": {
			v1: &config.ViewSetting{
				Columns: []string{"A", "B", "C"},
			},
			v2: &config.ViewSetting{
				Columns: []string{"A", "B", "C"},
			},
			e: true,
		},

		"order": {
			v1: &config.ViewSetting{
				Columns: []string{"C", "A", "B"},
			},
			v2: &config.ViewSetting{
				Columns: []string{"A", "B", "C"},
			},
		},

		"delta": {
			v1: &config.ViewSetting{
				Columns: []string{"A", "B", "C"},
			},
			v2: &config.ViewSetting{
				Columns: []string{"B"},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equalf(t, u.e, u.v1.Equals(u.v2), "%#v and %#v", u.v1, u.v2)
		})
	}
}
