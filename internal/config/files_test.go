// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"os"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func TestEnsureBenchmarkCfg(t *testing.T) {
	os.Setenv(config.K9sConfigDir, "/tmp/test-config")
	assert.NoError(t, config.InitLocs())
	defer assert.NoError(t, os.RemoveAll(config.K9sConfigDir))

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
