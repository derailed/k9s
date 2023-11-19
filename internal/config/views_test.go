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

	assert.Nil(t, cfg.Load("testdata/view_settings.yml"))
	assert.Equal(t, 1, len(cfg.K9s.Views))
	assert.Equal(t, 4, len(cfg.K9s.Views["v1/pods"].Columns))
}
