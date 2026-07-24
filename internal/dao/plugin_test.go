// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlugin(t *testing.T) {
	tmp := t.TempDir()
	globalPath := filepath.Join(tmp, "config", "plugins.yaml")
	contextPath := filepath.Join(tmp, "contexts", "plugins.yaml")

	require.NoError(t, os.MkdirAll(filepath.Dir(globalPath), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Dir(contextPath), 0o755))
	require.NoError(t, os.WriteFile(globalPath, []byte(`plugins:
  global-only:
    shortCut: Shift-G
    description: global
    scopes:
      - all
    command: kubectl
`), 0o600))
	require.NoError(t, os.WriteFile(contextPath, []byte(`plugins:
  context-only:
    shortCut: Shift-C
    description: context
    scopes:
      - pods
    command: kubectl
`), 0o600))

	origPluginsFile := config.AppPluginsFile
	origDataHome, origConfigHome, origDataDirs := xdg.DataHome, xdg.ConfigHome, xdg.DataDirs
	config.AppPluginsFile = globalPath
	xdg.DataHome = filepath.Join(tmp, "xdg-data")
	xdg.ConfigHome = filepath.Join(tmp, "xdg-config")
	xdg.DataDirs = nil
	t.Cleanup(func() {
		config.AppPluginsFile = origPluginsFile
		xdg.DataHome, xdg.ConfigHome, xdg.DataDirs = origDataHome, origConfigHome, origDataDirs
	})

	p := dao.NewPlugin(nil)
	ctx := context.WithValue(context.Background(), internal.KeyPath, contextPath)
	oo, err := p.List(ctx, "")

	require.NoError(t, err)
	require.Len(t, oo, 2)

	first, ok := oo[0].(render.PluginRes)
	require.True(t, ok)
	assert.Equal(t, "context-only", first.Name)

	second, ok := oo[1].(render.PluginRes)
	require.True(t, ok)
	assert.Equal(t, "global-only", second.Name)
}
