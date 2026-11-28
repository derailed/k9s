// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"log/slog"
	"os"
	"path/filepath"
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

type mockViewListener struct {
	ns       string
	calls    int
	settings []*config.ViewSetting
}

func (m *mockViewListener) ViewSettingsChanged(vs *config.ViewSetting) {
	m.calls++
	m.settings = append(m.settings, vs)
}

func (m *mockViewListener) GetNamespace() string {
	return m.ns
}

func TestCustomViewLoadMergeAndPrecedence(t *testing.T) {
	dir := t.TempDir()
	globalFile := filepath.Join(dir, "global-views.yaml")
	clusterFile := filepath.Join(dir, "cluster-views.yaml")
	contextFile := filepath.Join(dir, "context-views.yaml")

	require.NoError(t, os.WriteFile(globalFile, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - STATUS
  v1/services:
    columns:
      - NAME
      - TYPE
`), 0o600))

	require.NoError(t, os.WriteFile(clusterFile, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - NODE
  apps/v1/deployments:
    columns:
      - NAME
      - READY
`), 0o600))

	require.NoError(t, os.WriteFile(contextFile, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - IP
`), 0o600))

	cv := config.NewCustomView()
	lis := &mockViewListener{ns: "default"}
	cv.AddListener("v1/pods", lis)
	require.Equal(t, 1, lis.calls)

	require.NoError(t, cv.LoadFiles(globalFile, clusterFile, contextFile))

	assert.Equal(t, []string{"NAME", "IP"}, cv.Views["v1/pods"].Columns)
	assert.Equal(t, []string{"NAME", "READY"}, cv.Views["apps/v1/deployments"].Columns)
	assert.Equal(t, []string{"NAME", "TYPE"}, cv.Views["v1/services"].Columns)
	assert.Equal(t, 2, lis.calls, "LoadFiles should notify listeners once after merging all files")
	require.NotNil(t, lis.settings[len(lis.settings)-1])
	assert.Equal(t, []string{"NAME", "IP"}, lis.settings[len(lis.settings)-1].Columns)

	// Direct Load merges into existing views rather than replacing them.
	require.NoError(t, cv.Load(globalFile))
	require.NoError(t, cv.Load(clusterFile))
	assert.Equal(t, []string{"NAME", "NODE"}, cv.Views["v1/pods"].Columns)
	assert.Equal(t, []string{"NAME", "READY"}, cv.Views["apps/v1/deployments"].Columns)
	assert.Equal(t, []string{"NAME", "TYPE"}, cv.Views["v1/services"].Columns)

	// Resetting and loading only missing files must still notify listeners with nil.
	cv.Reset()
	require.NoError(t, cv.LoadFiles(filepath.Join(dir, "missing-1.yaml"), filepath.Join(dir, "missing-2.yaml")))
	assert.Empty(t, cv.Views)
	assert.Equal(t, 5, lis.calls)
	assert.Nil(t, lis.settings[len(lis.settings)-1])
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
