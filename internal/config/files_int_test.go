// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func Test_initXDGLocs(t *testing.T) {
	tmp, err := UserTmpDir()
	assert.NoError(t, err)

	os.Unsetenv("XDG_CONFIG_HOME")
	os.Unsetenv("XDG_CACHE_HOME")
	os.Unsetenv("XDG_STATE_HOME")
	os.Unsetenv("XDG_DATA_HOME")

	os.Setenv("XDG_CONFIG_HOME", filepath.Join(tmp, "k9s-xdg", "config"))
	os.Setenv("XDG_CACHE_HOME", filepath.Join(tmp, "k9s-xdg", "cache"))
	os.Setenv("XDG_STATE_HOME", filepath.Join(tmp, "k9s-xdg", "state"))
	os.Setenv("XDG_DATA_HOME", filepath.Join(tmp, "k9s-xdg", "data"))
	xdg.Reload()

	uu := map[string]struct {
		configDir          string
		configFile         string
		benchmarksDir      string
		contextsDir        string
		contextHotkeysFile string
		contextConfig      string
		dumpsDir           string
		benchDir           string
		hkFile             string
	}{
		"check-env": {
			configDir:          filepath.Join(tmp, "k9s-xdg", "config", "k9s"),
			configFile:         filepath.Join(tmp, "k9s-xdg", "config", "k9s", data.MainConfigFile),
			benchmarksDir:      filepath.Join(tmp, "k9s-xdg", "state", "k9s", "benchmarks"),
			contextsDir:        filepath.Join(tmp, "k9s-xdg", "data", "k9s", "clusters"),
			contextHotkeysFile: filepath.Join(tmp, "k9s-xdg", "data", "k9s", "clusters", "cl-1", "ct-1-1", "hotkeys.yaml"),
			contextConfig:      filepath.Join(tmp, "k9s-xdg", "data", "k9s", "clusters", "cl-1", "ct-1-1", data.MainConfigFile),
			dumpsDir:           filepath.Join(tmp, "k9s-xdg", "state", "k9s", "screen-dumps", "cl-1", "ct-1-1"),
			benchDir:           filepath.Join(tmp, "k9s-xdg", "state", "k9s", "benchmarks", "cl-1", "ct-1-1"),
			hkFile:             filepath.Join(tmp, "k9s-xdg", "config", "k9s", "hotkeys.yaml"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.NoError(t, initXDGLocs())
			assert.Equal(t, u.configDir, AppConfigDir)
			assert.Equal(t, u.configFile, AppConfigFile)
			assert.Equal(t, u.benchmarksDir, AppBenchmarksDir)
			assert.Equal(t, u.contextsDir, AppContextsDir)
			assert.Equal(t, u.contextHotkeysFile, AppContextHotkeysFile("cl-1", "ct-1-1"))
			assert.Equal(t, u.contextConfig, AppContextConfig("cl-1", "ct-1-1"))
			dir, err := DumpsDir("cl-1", "ct-1-1")
			assert.NoError(t, err)
			assert.Equal(t, u.dumpsDir, dir)
			bdir, err := EnsureBenchmarksDir("cl-1", "ct-1-1")
			assert.NoError(t, err)
			assert.Equal(t, u.benchDir, bdir)
			hk, err := EnsureHotkeysCfgFile()
			assert.NoError(t, err)
			assert.Equal(t, u.hkFile, hk)
		})
	}
}
