// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"path"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
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
			err:  "Additional property scoped is not allowed\nscopes is required\nAdditional property plugins is not allowed\ncommand is required\ndescription is required\nscopes is required\nshortCut is required\nAdditional property blah is not allowed\ncommand is required\ndescription is required\nscopes is required\nshortCut is required",
		},
	}

	dir, _ := os.Getwd()
	assert.NoError(t, os.Chdir("../.."))
	defer func() {
		assert.NoError(t, os.Chdir(dir))
	}()
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPlugins()
			err := p.Load(path.Join(dir, u.path), false)
			if err != nil {
				idx := strings.Index(err.Error(), ":")
				assert.Equal(t, u.err, err.Error()[idx+2:])
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

	dir, _ := os.Getwd()
	assert.NoError(t, os.Chdir("../.."))
	defer func() {
		assert.NoError(t, os.Chdir(dir))
	}()

	p := NewPlugins()
	assert.NoError(t, p.load(path.Join(dir, "testdata/plugins/plugins.yaml")))
	assert.NoError(t, p.loadDir(path.Join(dir, "/random/dir/not/exist")))

	assert.Equal(t, 1, len(p.Plugins))
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
			path: "internal/config/testdata/plugins/plugins.yaml",
			dir:  "internal/config/testdata/plugins/dir",
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

	dir, _ := os.Getwd()
	assert.NoError(t, os.Chdir("../.."))
	defer func() {
		assert.NoError(t, os.Chdir(dir))
	}()
	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			p := NewPlugins()
			assert.NoError(t, p.load(u.path))
			assert.NoError(t, p.loadDir(u.dir))
			assert.Equal(t, u.ee, p)
		})
	}
}
