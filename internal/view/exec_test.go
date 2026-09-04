// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestShellQuote(t *testing.T) {
	uu := map[string]struct {
		input string
		e     string
	}{
		"no-special-chars": {
			input: "simple",
			e:     "simple",
		},
		"path-with-spaces": {
			input: "/tmp/my editor/vi",
			e:     `"/tmp/my editor/vi"`,
		},
		"arg-with-spaces": {
			input: "set number",
			e:     `"set number"`,
		},
		"with-double-quotes": {
			input: `path"with"quotes`,
			e:     `"path\"with\"quotes"`,
		},
		"with-backslash": {
			input: `path\with\backslash`,
			e:     `"path\\with\\backslash"`,
		},
		"with-tabs": {
			input: "path\twith\ttabs",
			e:     "\"path\twith\ttabs\"",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, shellQuote(u.input))
		})
	}
}

func TestAsResource(t *testing.T) {
	t.Run("valid-limits", func(t *testing.T) {
		res := asResource(config.Limits{
			v1.ResourceCPU:    "500m",
			v1.ResourceMemory: "100Mi",
		})
		assert.Equal(t, "500m", res.Limits.Cpu().String())
		assert.Equal(t, "100Mi", res.Limits.Memory().String())
	})

	t.Run("empty-or-invalid-limits-do-not-panic", func(t *testing.T) {
		res := asResource(config.Limits{
			v1.ResourceCPU:    "",
			v1.ResourceMemory: "invalid-memory",
		})
		_, cpuFound := res.Limits[v1.ResourceCPU]
		assert.False(t, cpuFound)
		_, memFound := res.Limits[v1.ResourceMemory]
		assert.False(t, memFound)
	})
}
