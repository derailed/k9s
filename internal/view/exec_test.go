// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
