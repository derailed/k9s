// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ui_test

import (
	"context"
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
	s := newMockSynchronizer(cfg.Styles)
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

func newMockSynchronizer(styles *config.Styles) synchronizer {
	return synchronizer{logo: ui.NewLogo(styles)}
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
	s := newMockSynchronizer(cfg.Styles)
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

type channelSynchronizer struct {
	synchronizer
	draws chan func()
}

func (c channelSynchronizer) QueueUpdateDraw(fn func()) {
	c.draws <- fn
}

func TestRefreshCustomViewsClusterAndContextLayering(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, tmpDir)
	require.NoError(t, config.InitLocs())

	cfg := ui.Configurator{Config: mock.NewMockConfig(t)}
	// Restore AppContextsDir pointing to our temp dir since NewMockConfig overrides it.
	config.AppContextsDir = filepath.Join(tmpDir, "clusters")

	// Global views.yaml
	require.NoError(t, os.WriteFile(config.AppViewsFile, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - STATUS
  v1/services:
    columns:
      - NAME
      - TYPE
`), data.DefaultFileMod))

	// Cluster-level (context-agnostic) views.yaml for cl-1
	clusterViews := config.AppClusterViewsFile("cl-1")
	require.NoError(t, os.MkdirAll(filepath.Dir(clusterViews), data.DefaultDirMod))
	require.NoError(t, os.WriteFile(clusterViews, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - NODE
  apps/v1/deployments:
    columns:
      - NAME
      - READY
`), data.DefaultFileMod))

	// Context-level views.yaml only for ct-1-1 (ct-1-2 has none)
	ctxViews := config.AppContextViewsFile("cl-1", "ct-1-1")
	require.NoError(t, os.MkdirAll(filepath.Dir(ctxViews), data.DefaultDirMod))
	require.NoError(t, os.WriteFile(ctxViews, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - IP
`), data.DefaultFileMod))

	// Context ct-1-1: context > cluster > global
	_, err := cfg.Config.K9s.ActivateContext("ct-1-1")
	require.NoError(t, err)
	require.NoError(t, cfg.RefreshCustomViews())
	assert.Equal(t, []string{"NAME", "IP"}, cfg.CustomView().Views["v1/pods"].Columns)
	assert.Equal(t, []string{"NAME", "READY"}, cfg.CustomView().Views["apps/v1/deployments"].Columns)
	assert.Equal(t, []string{"NAME", "TYPE"}, cfg.CustomView().Views["v1/services"].Columns)

	// Different context ct-1-2 on the SAME cluster cl-1 shares cluster-level views.yaml
	_, err = cfg.Config.K9s.ActivateContext("ct-1-2")
	require.NoError(t, err)
	require.NoError(t, cfg.RefreshCustomViews())
	assert.Equal(t, []string{"NAME", "NODE"}, cfg.CustomView().Views["v1/pods"].Columns)
	assert.Equal(t, []string{"NAME", "READY"}, cfg.CustomView().Views["apps/v1/deployments"].Columns)
	assert.Equal(t, []string{"NAME", "TYPE"}, cfg.CustomView().Views["v1/services"].Columns)

	// Switching to a different cluster without cluster views falls back to global only
	_, err = cfg.Config.K9s.ActivateContext("fred-blee")
	require.NoError(t, err)
	require.NoError(t, cfg.RefreshCustomViews())
	assert.Equal(t, []string{"NAME", "STATUS"}, cfg.CustomView().Views["v1/pods"].Columns)
	assert.Equal(t, []string{"NAME", "TYPE"}, cfg.CustomView().Views["v1/services"].Columns)
	assert.NotContains(t, cfg.CustomView().Views, "apps/v1/deployments")
}

func TestCustomViewsWatcherClusterLevelWithoutGlobal(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv(config.K9sEnvConfigDir, tmpDir)
	require.NoError(t, config.InitLocs())

	cfg := ui.Configurator{
		Config: mock.NewMockConfig(t),
		Styles: config.NewStyles(),
	}
	config.AppContextsDir = filepath.Join(tmpDir, "clusters")

	_, err := cfg.Config.K9s.ActivateContext("ct-1-2")
	require.NoError(t, err)

	// Only cluster-level views.yaml exists (no global AppViewsFile).
	clusterViews := config.AppClusterViewsFile("cl-1")
	require.NoError(t, os.MkdirAll(filepath.Dir(clusterViews), data.DefaultDirMod))
	require.NoError(t, os.WriteFile(clusterViews, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - NODE
`), data.DefaultFileMod))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	syncer := channelSynchronizer{
		synchronizer: newMockSynchronizer(cfg.Styles),
		draws:        make(chan func(), 16),
	}
	require.NoError(t, cfg.CustomViewsWatcher(ctx, syncer))
	assert.Equal(t, []string{"NAME", "NODE"}, cfg.CustomView().Views["v1/pods"].Columns)

	// Updating cluster-level views.yaml triggers live reload via watcher.
	require.NoError(t, os.WriteFile(clusterViews, []byte(`views:
  v1/pods:
    columns:
      - NAME
      - IP
      - AGE
`), data.DefaultFileMod))

	deadline := time.After(3 * time.Second)
	for {
		select {
		case fn := <-syncer.draws:
			fn()
			if assert.ObjectsAreEqual([]string{"NAME", "IP", "AGE"}, cfg.CustomView().Views["v1/pods"].Columns) {
				return
			}
		case <-deadline:
			t.Fatalf("timed out waiting for cluster views reload; got %v", cfg.CustomView().Views["v1/pods"].Columns)
		}
	}
}
