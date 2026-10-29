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

func TestSplitWindowsCommand(t *testing.T) {
	uu := map[string]struct {
		input string
		e     []string
	}{
		"bare": {
			input: "notepad",
			e:     []string{"notepad"},
		},
		"with-args": {
			input: "code -w",
			e:     []string{"code", "-w"},
		},
		"backslash-path": {
			input: `C:\Windows\notepad.exe`,
			e:     []string{`C:\Windows\notepad.exe`},
		},
		"quoted-path-with-spaces": {
			input: `"C:\Program Files\Notepad++\notepad++.exe" -multiInst  -nosession`,
			e:     []string{`C:\Program Files\Notepad++\notepad++.exe`, "-multiInst", "-nosession"},
		},
		"unc-path": {
			input: `\server\share\vi.exe`,
			e:     []string{`\server\share\vi.exe`},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, splitWindowsCommand(u.input))
		})
	}
}

func TestKubeEditor(t *testing.T) {
	uu := map[string]struct {
		input []string
		e     string
	}{
		"bare-backslash-path": {
			input: []string{`C:\Windows\notepad.exe`},
			e:     `C:\Windows\notepad.exe`,
		},
		"path-with-spaces": {
			input: []string{"/tmp/my editor/vi"},
			e:     `"/tmp/my editor/vi"`,
		},
		"with-args": {
			input: []string{`C:\bin\code.cmd`, "-w"},
			e:     `"C:\\bin\\code.cmd" -w`,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, kubeEditor(u.input))
		})
	}
}
