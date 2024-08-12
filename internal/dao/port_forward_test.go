// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestBenchForConfig(t *testing.T) {
	uu := map[string]struct {
		file, key string
		spec      config.BenchConfig
	}{
		"no_file": {file: "", key: "", spec: config.DefaultBenchSpec()},
		"spec": {file: "testdata/benchspec.yaml", key: "default/nginx-123-456|nginx", spec: config.BenchConfig{
			C: 2,
			N: 3000,
			HTTP: config.HTTP{
				Method: "GET",
				Path:   "/",
			},
		}},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.NotNil(t, u.spec, dao.BenchConfigFor(u.file, u.key))
		})
	}
}
