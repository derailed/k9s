// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginLoad(t *testing.T) {
	uu := map[string]struct {
		path string
		err  string
		ee   Plugins
	}{
		"snippet": {
			path: "testdata/plugins/dir/snippet.1.yaml",
			ee: Plugins{
				Plugins: plugins{
					"snippet.1": Plugin{
						Scopes:          []string{"po", "dp"},
						Args:            []string{"-n", "$NAMESPACE", "-boolean"},
						ShortCut:        "shift-s",
						Description:     "blee",
						Command:         "duh",
						Confirm:         true,
						OverwriteOutput: true,
					},
				},
			},
		},

		"multi-snippets": {
			path: "testdata/plugins/dir/snippet.multi.yaml",
			ee: Plugins{
				Plugins: plugins{
					"crapola": Plugin{
						ShortCut:    "Shift-1",
						Command:     "crapola",
						Description: "crapola",
						Scopes:      []string{"pods"},
					},
					"bozo": Plugin{
						ShortCut:    "Shift-2",
						Description: "bozo",
						Command:     "bozo",
						Scopes:      []string{"pods", "svc"},
					},
				},
			},
		},

		"full": {
			path: "testdata/plugins/plugins.yaml",
			ee: Plugins{
				Plugins: plugins{
					"blah": Plugin{
						Scopes:      []string{"po", "dp"},
						Args:        []string{"-n", "$NAMESPACE", "-boolean"},
						ShortCut:    "shift-s",
						Description: "blee",
						Command:     "duh",
						Confirm:     true,
					},
				},
			},
		},

		"toast-no-file": {
			path: "testdata/plugins/plugins-bozo.yaml",
			ee:   NewPlugins(),
		},

		"toast-invalid": {
			path: "testdata/plugins/plugins-toast.yaml",
			ee:   NewPlugins(),
			err:  "plugin validation failed for testdata/plugins/plugins-toast.yaml: scopes is required\nAdditional property plugins is not allowed\ncommand is required\ndescription is required\nscopes is required\nshortCut is required\ncommand is required\ndescription is required\nscopes is required\nshortCut is required",
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPlugins()
			err := p.Load(u.path, false)
			if err != nil {
				assert.Equal(t, u.err, err.Error())
			}
			assert.Equal(t, u.ee, p)
		})
	}
}

func TestSinglePluginFileLoad(t *testing.T) {
	e := Plugin{
		Scopes:      []string{"po", "dp"},
		Args:        []string{"-n", "$NAMESPACE", "-boolean"},
		ShortCut:    "shift-s",
		Description: "blee",
		Command:     "duh",
		Confirm:     true,
	}

	p := NewPlugins()
	require.NoError(t, p.load("testdata/plugins/plugins.yaml"))
	require.NoError(t, p.loadDir("/random/dir/not/exist"))

	assert.Len(t, p.Plugins, 1)
	v, ok := p.Plugins["blah"]

	assert.True(t, ok)
	assert.ObjectsAreEqual(e, v)
}

func TestMultiplePluginFilesLoad(t *testing.T) {
	uu := map[string]struct {
		path string
		dir  string
		ee   Plugins
	}{
		"empty": {
			path: "testdata/plugins/plugins.yaml",
			dir:  "testdata/plugins/dir",
			ee: Plugins{
				Plugins: plugins{
					"blah": {
						Scopes:      []string{"po", "dp"},
						Args:        []string{"-n", "$NAMESPACE", "-boolean"},
						ShortCut:    "shift-s",
						Description: "blee",
						Command:     "duh",
						Confirm:     true,
					},
					"snippet.1": {
						ShortCut:        "shift-s",
						Command:         "duh",
						Scopes:          []string{"po", "dp"},
						Args:            []string{"-n", "$NAMESPACE", "-boolean"},
						Description:     "blee",
						Confirm:         true,
						OverwriteOutput: true,
					},
					"snippet.2": {
						Scopes:      []string{"svc", "ing"},
						Args:        []string{"-n", "$NAMESPACE", "-oyaml"},
						ShortCut:    "shift-r",
						Description: "bla",
						Command:     "duha",
						Background:  true,
					},
					"crapola": {
						Scopes:      []string{"pods"},
						Command:     "crapola",
						Description: "crapola",
						ShortCut:    "Shift-1",
					},
					"bozo": {
						Scopes:      []string{"pods", "svc"},
						Command:     "bozo",
						Description: "bozo",
						ShortCut:    "Shift-2",
					},
				},
			},
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPlugins()
			require.NoError(t, p.load(u.path))
			require.NoError(t, p.loadDir(u.dir))
			assert.Equal(t, u.ee, p)
		})
	}
}

func TestPluginLoadSymlink(t *testing.T) {
	tmp := t.TempDir()

	linkFile := filepath.Join(tmp, "plugins-symlink.yaml")
	wd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Symlink(filepath.Join(wd, "testdata", "plugins", "plugins.yaml"), linkFile))

	linkDir := filepath.Join(tmp, "plugins-dir-symlink")
	require.NoError(t, os.Symlink(filepath.Join(wd, "testdata", "plugins", "dir"), linkDir))

	// Add a symlink with an infinite loop
	loopDir := filepath.Join(tmp, "loop")
	require.NoError(t, os.Mkdir(loopDir, 0o755))
	require.NoError(t, os.Symlink(loopDir, filepath.Join(loopDir, "self")))

	p := NewPlugins()
	require.NoError(t, p.loadDir(tmp))

	ee := Plugins{
		Plugins: plugins{
			"blah": Plugin{
				Scopes:      []string{"po", "dp"},
				Args:        []string{"-n", "$NAMESPACE", "-boolean"},
				ShortCut:    "shift-s",
				Description: "blee",
				Command:     "duh",
				Confirm:     true,
			},
			"snippet.1": {
				ShortCut:        "shift-s",
				Command:         "duh",
				Scopes:          []string{"po", "dp"},
				Args:            []string{"-n", "$NAMESPACE", "-boolean"},
				Description:     "blee",
				Confirm:         true,
				OverwriteOutput: true,
			},
			"snippet.2": {
				Scopes:      []string{"svc", "ing"},
				Args:        []string{"-n", "$NAMESPACE", "-oyaml"},
				ShortCut:    "shift-r",
				Description: "bla",
				Command:     "duha",
				Background:  true,
			},
			"crapola": {
				Scopes:      []string{"pods"},
				Command:     "crapola",
				Description: "crapola",
				ShortCut:    "Shift-1",
			},
			"bozo": {
				Scopes:      []string{"pods", "svc"},
				Command:     "bozo",
				Description: "bozo",
				ShortCut:    "Shift-2",
			},
		},
	}

	assert.Equal(t, ee, p)
}
