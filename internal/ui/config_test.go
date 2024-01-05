// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestSkinnedContext(t *testing.T) {
	os.Setenv(config.K9sEnvConfigDir, "/tmp/test-config")

	assert.NoError(t, config.InitLocs())
	defer assert.NoError(t, os.RemoveAll(config.K9sEnvConfigDir))

	sf := filepath.Join("..", "config", "testdata", "black_and_wtf.yaml")
	raw, err := os.ReadFile(sf)
	assert.NoError(t, err)
	tf := filepath.Join(config.AppSkinsDir, "black_and_wtf.yaml")
	assert.NoError(t, os.WriteFile(tf, raw, data.DefaultFileMod))

	var cfg ui.Configurator
	cfg.Config = mock.NewMockConfig()
	cl, ct := "cl-1", "ct-1"
	flags := genericclioptions.ConfigFlags{
		ClusterName: &cl,
		Context:     &ct,
	}

	cfg.Config.K9s = config.NewK9s(
		mock.NewMockConnection(),
		mock.NewMockKubeSettings(&flags))
	_, err = cfg.Config.K9s.ActivateContext("ct-1-1")
	assert.NoError(t, err)
	cfg.Config.K9s.UI = config.UI{Skin: "black_and_wtf"}
	cfg.RefreshStyles()

	assert.True(t, cfg.HasSkin())
	assert.Equal(t, tcell.ColorGhostWhite.TrueColor(), render.StdColor)
	assert.Equal(t, tcell.ColorWhiteSmoke.TrueColor(), render.ErrColor)
}

func TestBenchConfig(t *testing.T) {
	os.Setenv(config.K9sEnvConfigDir, "/tmp/test-config")
	assert.NoError(t, config.InitLocs())
	defer assert.NoError(t, os.RemoveAll(config.K9sEnvConfigDir))

	bc, error := config.EnsureBenchmarksCfgFile("cl-1", "ct-1")
	assert.NoError(t, error)
	assert.Equal(t, "/tmp/test-config/clusters/cl-1/ct-1/benchmarks.yaml", bc)
}
