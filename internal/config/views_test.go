// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"log/slog"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
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
			key:  "v1/pods",
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

			assert.NoError(t, cfg.Load(u.path))
			assert.Equal(t, u.e, cfg.Views[u.key].Columns)
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
