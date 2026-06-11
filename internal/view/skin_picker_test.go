// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestInstalledSkinNames(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(dir, "subdir"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "zeta.yaml"), []byte("zeta"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "alpha.yaml"), []byte("alpha"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "beta.yml"), []byte("beta"), 0o600))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("notes"), 0o600))

	got, err := installedSkinNames(dir)
	require.NoError(t, err)
	assert.Equal(t, []string{"alpha", "zeta"}, got)
}

func TestInstalledSkinNamesEmpty(t *testing.T) {
	got, err := installedSkinNames(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestApplySkinSelectionGlobal(t *testing.T) {
	app := newSkinTestApp(t)

	writeGlobalConfigFixture(t, config.AppConfigFile)
	require.NoError(t, app.Config.Load(config.AppConfigFile, true))
	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")

	result, err := app.applySkinSelection(skinScopeGlobal, "black-and-wtf")
	require.NoError(t, err)

	assert.Equal(t, "black-and-wtf", app.Config.K9s.UI.Skin)
	assert.True(t, app.HasSkin())
	assert.Empty(t, result.warnings)
	assert.Equal(t, tcell.ColorGhostWhite.TrueColor(), model1.StdColor)

	var stored config.Config
	readYAML(t, config.AppConfigFile, &stored)
	require.NotNil(t, stored.K9s)
	assert.Equal(t, "black-and-wtf", stored.K9s.UI.Skin)
}

func TestApplySkinSelectionContext(t *testing.T) {
	app := newSkinTestApp(t)

	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")

	result, err := app.applySkinSelection(skinScopeContext, "black-and-wtf")
	require.NoError(t, err)

	ct, err := app.Config.CurrentContext()
	require.NoError(t, err)
	assert.Equal(t, "black-and-wtf", ct.Skin)
	assert.True(t, app.HasSkin())
	assert.Empty(t, result.warnings)

	var stored data.Config
	readYAML(t, config.AppContextConfig("cl-1", "ct-1-1"), &stored)
	require.NotNil(t, stored.Context)
	assert.Equal(t, "black-and-wtf", stored.Context.Skin)
}

func TestApplySkinSelectionClearsGlobalSkin(t *testing.T) {
	app := newSkinTestApp(t)

	writeGlobalConfigFixture(t, config.AppConfigFile)
	require.NoError(t, app.Config.Load(config.AppConfigFile, true))
	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")

	_, err := app.applySkinSelection(skinScopeGlobal, "black-and-wtf")
	require.NoError(t, err)

	result, err := app.applySkinSelection(skinScopeGlobal, "")
	require.NoError(t, err)

	assert.Empty(t, app.Config.K9s.UI.Skin)
	assert.False(t, app.HasSkin())
	assert.Contains(t, result.notice(), "Default skin restored")

	var stored config.Config
	readYAML(t, config.AppConfigFile, &stored)
	require.NotNil(t, stored.K9s)
	assert.Empty(t, stored.K9s.UI.Skin)
}

func TestApplySkinSelectionClearsContextSkin(t *testing.T) {
	app := newSkinTestApp(t)

	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")

	_, err := app.applySkinSelection(skinScopeContext, "black-and-wtf")
	require.NoError(t, err)

	result, err := app.applySkinSelection(skinScopeContext, "")
	require.NoError(t, err)

	ct, err := app.Config.CurrentContext()
	require.NoError(t, err)
	assert.Empty(t, ct.Skin)
	assert.False(t, app.HasSkin())
	assert.Contains(t, result.notice(), "Default skin restored")

	var stored data.Config
	readYAML(t, config.AppContextConfig("cl-1", "ct-1-1"), &stored)
	require.NotNil(t, stored.Context)
	assert.Empty(t, stored.Context.Skin)
}

func TestApplySkinSelectionWarnsOnEnvOverride(t *testing.T) {
	app := newSkinTestApp(t)
	t.Setenv("K9S_SKIN", "env-skin")

	writeGlobalConfigFixture(t, config.AppConfigFile)
	require.NoError(t, app.Config.Load(config.AppConfigFile, true))
	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")

	result, err := app.applySkinSelection(skinScopeGlobal, "black-and-wtf")
	require.NoError(t, err)
	assert.Contains(t, strings.Join(result.warnings, "\n"), "K9S_SKIN")
}

func TestApplySkinSelectionWarnsOnContextOverride(t *testing.T) {
	app := newSkinTestApp(t)

	writeGlobalConfigFixture(t, config.AppConfigFile)
	require.NoError(t, app.Config.Load(config.AppConfigFile, true))
	writeSkinFixture(t, config.AppSkinsDir, "black-and-wtf")
	writeSkinFixture(t, config.AppSkinsDir, "ctx-override")

	ct, err := app.Config.CurrentContext()
	require.NoError(t, err)
	ct.Skin = "ctx-override"

	result, err := app.applySkinSelection(skinScopeGlobal, "black-and-wtf")
	require.NoError(t, err)
	assert.Contains(t, strings.Join(result.warnings, "\n"), "context-specific skin")
}

func TestHelpShowGeneralIncludesSkins(t *testing.T) {
	h := &Help{}

	found := false
	for _, hint := range h.showGeneral() {
		if hint.Mnemonic == "Shift-t" && hint.Description == "Skins" {
			found = true
			break
		}
	}

	assert.True(t, found)
}

func newSkinTestApp(t testing.TB) *App {
	t.Helper()

	oldConfigFile := config.AppConfigFile
	oldSkinsDir := config.AppSkinsDir
	oldContextsDir := config.AppContextsDir
	tmp := t.TempDir()
	config.AppConfigFile = filepath.Join(tmp, "config.yaml")
	config.AppSkinsDir = filepath.Join(tmp, "skins")
	contextsDir := filepath.Join(tmp, "clusters")
	require.NoError(t, os.MkdirAll(config.AppSkinsDir, 0o700))
	require.NoError(t, os.MkdirAll(contextsDir, 0o700))
	t.Cleanup(func() {
		config.AppConfigFile = oldConfigFile
		config.AppSkinsDir = oldSkinsDir
		config.AppContextsDir = oldContextsDir
	})

	cfg := mock.NewMockConfig(t)
	config.AppContextsDir = contextsDir
	_, err := cfg.K9s.ActivateContext("ct-1-1")
	require.NoError(t, err)

	return NewApp(cfg)
}

func writeGlobalConfigFixture(t testing.TB, path string) {
	t.Helper()

	bb, err := os.ReadFile(filepath.Join("..", "config", "testdata", "configs", "default.yaml"))
	require.NoError(t, err)
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o700))
	require.NoError(t, os.WriteFile(path, bb, 0o600))
}

func writeSkinFixture(t testing.TB, dir, name string) {
	t.Helper()

	bb, err := os.ReadFile(filepath.Join("..", "config", "testdata", "skins", "black-and-wtf.yaml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, name+".yaml"), bb, 0o600))
}

func readYAML(t testing.TB, path string, target any) {
	t.Helper()

	bb, err := os.ReadFile(path)
	require.NoError(t, err)
	require.NoError(t, yaml.Unmarshal(bb, target))
}
