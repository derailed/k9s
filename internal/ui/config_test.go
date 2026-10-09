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
	cfg.Config = mock.NewMockConfig(t)
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
	cfg.Config.K9s.UI = config.UI{Skin: "black-and-wtf", Logo: "CUSTOM"}
	cfg.Styles = config.NewStyles()
	s := newMockSynchronizer()
	s.logo = ui.NewLogo(cfg.Styles)
	cfg.RefreshStyles(s)
	assert.True(t, cfg.HasSkin())
	assert.Equal(t, tcell.ColorGhostWhite.TrueColor(), model1.StdColor)
	assert.Equal(t, tcell.ColorWhiteSmoke.TrueColor(), model1.ErrColor)
	screen := tcell.NewSimulationScreen("UTF-8")
	require.NoError(t, screen.Init())
	defer screen.Fini()
	s.logo.SetRect(0, 0, ui.LogoWidth, ui.LogoArtHeight+1)
	s.logo.Draw(screen)
	r, _, style, _ := screen.GetContent(0, 0)
	assert.Equal(t, 'C', r)
	fg, _, _ := style.Decompose()
	assert.Equal(t, tcell.ColorWhite.TrueColor(), fg)
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

type synchronizer struct{ logo *ui.Logo }

func newMockSynchronizer() synchronizer {
	return synchronizer{logo: ui.NewLogo(config.NewStyles())}
}

func (synchronizer) Flash() *model.Flash {
	return model.NewFlash(100 * time.Millisecond)
}
func (s synchronizer) Logo() *ui.Logo       { return s.logo }
func (synchronizer) UpdateClusterInfo()     {}
func (synchronizer) QueueUpdateDraw(func()) {}
func (synchronizer) QueueUpdate(func())     {}

func TestRefreshStylesLogo(t *testing.T) {
	cfg := ui.Configurator{Config: mock.NewMockConfig(t), Styles: config.NewStyles()}
	s := synchronizer{logo: ui.NewLogo(cfg.Styles)}
	cfg.Config.K9s.UI.Logo = "GLOBAL"
	cfg.RefreshStyles(s)
	assert.Contains(t, s.Logo().Logo().GetText(false), "GLOBAL")
	ct, err := cfg.Config.K9s.ActivateContext("ct-1-1")
	require.NoError(t, err)
	ct.Logo = "CONTEXT"
	cfg.RefreshStyles(s)
	assert.Contains(t, s.Logo().Logo().GetText(false), "CONTEXT")
	cfg.RefreshStyles(s)
	assert.Contains(t, s.Logo().Logo().GetText(false), "CONTEXT")
	ct.Logo = " \n\t"
	cfg.RefreshStyles(s)
	assert.Contains(t, s.Logo().Logo().GetText(false), "GLOBAL")
}
