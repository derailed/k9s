// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplitCommandLine(t *testing.T) {
	uu := map[string]struct {
		in string
		e  []string
	}{
		"editor with quoted path": {
			in: `"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code" -w`,
			e:  []string{"/Applications/Visual Studio Code.app/Contents/Resources/app/bin/code", "-w"},
		},
		"pipe with quoted jq program": {
			in: `jq -r '.items[] | .metadata.name'`,
			e:  []string{"jq", "-r", ".items[] | .metadata.name"},
		},
	}

	for name := range uu {
		u := uu[name]
		t.Run(name, func(t *testing.T) {
			got, err := splitCommandLine(u.in)
			require.NoError(t, err)
			assert.Equal(t, u.e, got)
		})
	}
}
