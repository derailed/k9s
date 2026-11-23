// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCompletePath(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(root, "compose", "prod"), 0o750))
	require.NoError(t, os.Mkdir(filepath.Join(root, ".hidden"), 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(root, "compose.yaml"), nil, 0o600))

	sep := string(os.PathSeparator)
	tests := []struct {
		name string
		text string
		want []string
	}{
		{
			name: "empty input",
			text: "",
			want: nil,
		},
		{
			name: "no match",
			text: filepath.Join(root, "zzz"),
			want: []string{},
		},
		{
			name: "prefix match",
			text: filepath.Join(root, "com"),
			want: []string{
				filepath.Join(root, "compose") + sep,
				filepath.Join(root, "compose.yaml"),
			},
		},
		{
			name: "trailing separator lists directory",
			text: filepath.Join(root, "compose") + sep,
			want: []string{filepath.Join(root, "compose", "prod") + sep},
		},
		{
			name: "hidden entries only on request",
			text: root + sep + ".",
			want: []string{filepath.Join(root, ".hidden") + sep},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, completePath(tt.text))
		})
	}
}

func TestCompletePathTilde(t *testing.T) {
	root := t.TempDir()
	require.NoError(t, os.Mkdir(filepath.Join(root, "compose"), 0o750))
	t.Setenv("HOME", root)
	t.Setenv("USERPROFILE", root)

	sep := string(os.PathSeparator)
	assert.Equal(t,
		[]string{"~" + sep + "compose" + sep},
		completePath("~"+sep+"com"),
	)
}
