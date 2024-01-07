// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func TestInitLogLoc(t *testing.T) {
	tmp, err := config.UserTmpDir()
	assert.NoError(t, err)

	uu := map[string]struct {
		dir string
		e   string
	}{
		"log-env": {
			dir: "/tmp/test/k9s/logs",
			e:   "/tmp/test/k9s/logs/k9s.log",
		},
		"xdg-env": {
			dir: "/tmp/test/xdg-state",
			e:   "/tmp/test/xdg-state/k9s/k9s.log",
		},
		"cfg-env": {
			dir: "/tmp/test/k9s-test",
			e:   filepath.Join(tmp, "k9s.log"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			os.Unsetenv(config.K9sEnvLogsDir)
			os.Unsetenv("XDG_STATE_HOME")
			os.Unsetenv(config.K9sEnvConfigDir)
			switch k {
			case "log-env":
				os.Setenv(config.K9sEnvLogsDir, u.dir)
			case "xdg-env":
				os.Setenv("XDG_STATE_HOME", u.dir)
				xdg.Reload()
			case "cfg-env":
				os.Setenv(config.K9sEnvConfigDir, u.dir)
			}
			err := config.InitLogLoc()
			assert.NoError(t, err)
			assert.Equal(t, u.e, config.AppLogFile)
			assert.NoError(t, os.RemoveAll(config.AppLogFile))
		})
	}
}

func TestEnsureBenchmarkCfg(t *testing.T) {
	os.Setenv(config.K9sEnvConfigDir, "/tmp/test-config")
	assert.NoError(t, config.InitLocs())
	defer assert.NoError(t, os.RemoveAll("/tmp/test-config"))

	assert.NoError(t, data.EnsureFullPath("/tmp/test-config/clusters/cl-1/ct-2", data.DefaultDirMod))
	assert.NoError(t, os.WriteFile("/tmp/test-config/clusters/cl-1/ct-2/benchmarks.yaml", []byte{}, data.DefaultFileMod))

	uu := map[string]struct {
		cluster, context string
		f, e             string
	}{
		"not-exist": {
			cluster: "cl-1",
			context: "ct-1",
			f:       "/tmp/test-config/clusters/cl-1/ct-1/benchmarks.yaml",
			e:       "benchmarks:\n  defaults:\n    concurrency: 2\n    requests: 200",
		},
		"exist": {
			cluster: "cl-1",
			context: "ct-2",
			f:       "/tmp/test-config/clusters/cl-1/ct-2/benchmarks.yaml",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			f, err := config.EnsureBenchmarksCfgFile(u.cluster, u.context)
			assert.NoError(t, err)
			assert.Equal(t, u.f, f)
			bb, err := os.ReadFile(f)
			assert.NoError(t, err)
			assert.Equal(t, u.e, string(bb))
		})
	}
}

func TestSkinFileFromName(t *testing.T) {
	config.AppSkinsDir = "/tmp/k9s-test/skins"
	defer assert.NoError(t, os.RemoveAll("/tmp/k9s-test/skins"))

	uu := map[string]struct {
		n string
		e string
	}{
		"empty": {
			e: "/tmp/k9s-test/skins/stock.yaml",
		},
		"happy": {
			n: "fred-blee",
			e: "/tmp/k9s-test/skins/fred-blee.yaml",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, config.SkinFileFromName(u.n))
		})
	}
}
