// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestViewSettingsLoad(t *testing.T) {
	cfg := config.NewCustomView()

	assert.Nil(t, cfg.Load("testdata/views/views.yaml"))
	assert.Equal(t, 1, len(cfg.Views))
	assert.Equal(t, 4, len(cfg.Views["v1/pods"].Columns))
}

func TestViewSetting_Equals(t *testing.T) {
	tests := []struct {
		v1, v2 *config.ViewSetting
		equals bool
	}{
		{nil, nil, true},
		{&config.ViewSetting{}, nil, false},
		{nil, &config.ViewSetting{}, false},
		{&config.ViewSetting{}, &config.ViewSetting{}, true},
		{&config.ViewSetting{Columns: []string{"A"}}, &config.ViewSetting{}, false},
		{&config.ViewSetting{Columns: []string{"A"}}, &config.ViewSetting{Columns: []string{"A"}}, true},
		{&config.ViewSetting{Columns: []string{"A"}}, &config.ViewSetting{Columns: []string{"B"}}, false},
		{&config.ViewSetting{SortColumn: "A"}, &config.ViewSetting{SortColumn: "B"}, false},
		{&config.ViewSetting{SortColumn: "A"}, &config.ViewSetting{SortColumn: "A"}, true},
	}

	for _, tt := range tests {
		assert.Equalf(t, tt.equals, tt.v1.Equals(tt.v2), "%#v and %#v", tt.v1, tt.v2)
	}
}
