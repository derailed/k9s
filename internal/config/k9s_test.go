// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/stretchr/testify/assert"
)

func TestGetScreenDumpDir(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	assert.Equal(t, "/tmp", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirOverride(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	cfg.K9s.OverrideScreenDumpDir("/override")
	assert.Equal(t, "/override", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirOverrideEmpty(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s.yaml"))
	cfg.K9s.OverrideScreenDumpDir("")
	assert.Equal(t, "/tmp", cfg.K9s.GetScreenDumpDir())
}

func TestGetScreenDumpDirEmpty(t *testing.T) {
	cfg := mock.NewMockConfig()

	assert.Nil(t, cfg.Load("testdata/k9s1.yaml"))
	cfg.K9s.OverrideScreenDumpDir("")
	assert.Equal(t, config.AppDumpsDir, cfg.K9s.GetScreenDumpDir())
}
