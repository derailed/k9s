// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func TestSkinnedContext(t *testing.T) {
	require.NoError(t, os.Setenv(config.K9sEnvConfigDir, "/tmp/k9s-test"))
	require.NoError(t, config.InitLocs())
	defer require.NoError(t, os.RemoveAll(config.K9sEnvConfigDir))

	sf := filepath.Join("..", "config", "testdata", "skins", "black-and-wtf.yaml")
	raw, err := os.ReadFile(sf)
	require.NoError(t, err)
	tf := filepath.Join(config.AppSkinsDir, "black-and-wtf.yaml")
	require.NoError(t, os.WriteFile(tf, raw, data.DefaultFileMod))

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
	require.NoError(t, err)
	cfg.Config.K9s.UI = config.UI{Skin: "black-and-wtf"}
	cfg.RefreshStyles(newMockSynchronizer())
	assert.True(t, cfg.HasSkin())
	assert.Equal(t, tcell.ColorGhostWhite.TrueColor(), model1.StdColor)
	assert.Equal(t, tcell.ColorWhiteSmoke.TrueColor(), model1.ErrColor)
}

func TestBenchConfig(t *testing.T) {
	require.NoError(t, os.Setenv(config.K9sEnvConfigDir, "/tmp/test-config"))
	require.NoError(t, config.InitLocs())
	defer require.NoError(t, os.RemoveAll(config.K9sEnvConfigDir))

	bc, err := config.EnsureBenchmarksCfgFile("cl-1", "ct-1")
	require.NoError(t, err)
	assert.Equal(t, "/tmp/test-config/clusters/cl-1/ct-1/benchmarks.yaml", bc)
}

// Helpers...

type synchronizer struct{}

func newMockSynchronizer() synchronizer {
	return synchronizer{}
}

func (synchronizer) Flash() *model.Flash {
	return model.NewFlash(100 * time.Millisecond)
}
func (synchronizer) Logo() *ui.Logo         { return nil }
func (synchronizer) UpdateClusterInfo()     {}
func (synchronizer) QueueUpdateDraw(func()) {}
func (synchronizer) QueueUpdate(func())     {}
