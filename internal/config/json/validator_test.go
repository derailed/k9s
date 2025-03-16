// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package json_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/derailed/k9s/internal/config/json"
	"github.com/stretchr/testify/assert"
)

func TestValidatePluginSnippet(t *testing.T) {
	plugPath := "testdata/plugins/snippet.yaml"
	bb, err := os.ReadFile(plugPath)
	assert.NoError(t, err)

	p := json.NewValidator()
	assert.NoError(t, p.Validate(json.PluginSchema, bb), plugPath)
}

func TestValidatePlugins(t *testing.T) {
	uu := map[string]struct {
		path, schema string
		err          string
	}{
		"cool": {
			path:   "testdata/plugins/cool.yaml",
			schema: json.PluginsSchema,
		},
		"toast": {
			path:   "testdata/plugins/toast.yaml",
			schema: json.PluginsSchema,
			err:    "scopes is required\nshortCut is required",
		},
		"cool-snippet": {
			path:   "testdata/plugins/snippet.yaml",
			schema: json.PluginSchema,
		},
		"cool-snippets": {
			path:   "testdata/plugins/snippets.yaml",
			schema: json.PluginMultiSchema,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.path)
			assert.NoError(t, err)
			v := json.NewValidator()
			if err := v.Validate(u.schema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestValidatePluginDir(t *testing.T) {
	plugDir := "../../../plugins"
	ee, err := os.ReadDir(plugDir)
	assert.NoError(t, err)
	for _, e := range ee {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		if ext == ".md" {
			continue
		}
		assert.True(t, ext == ".yaml", fmt.Sprintf("expected yaml file: %q", e.Name()))
		assert.False(t, strings.Contains(e.Name(), "_"), fmt.Sprintf("underscore in: %q", e.Name()))
		bb, err := os.ReadFile(filepath.Join(plugDir, e.Name()))
		assert.NoError(t, err)

		p := json.NewValidator()
		assert.NoError(t, p.Validate(json.PluginsSchema, bb), e.Name())
	}
}

func TestValidateSkinDir(t *testing.T) {
	skinDir := "../../../skins"
	ee, err := os.ReadDir(skinDir)
	assert.NoError(t, err)
	p := json.NewValidator()
	for _, e := range ee {
		if e.IsDir() {
			continue
		}
		ext := filepath.Ext(e.Name())
		assert.True(t, ext == ".yaml", fmt.Sprintf("expected yaml file: %q", e.Name()))
		assert.True(t, !strings.Contains(e.Name(), "_"), fmt.Sprintf("underscore in: %q", e.Name()))
		bb, err := os.ReadFile(filepath.Join(skinDir, e.Name()))
		assert.NoError(t, err)
		assert.NoError(t, p.Validate(json.SkinSchema, bb), e.Name())
	}
}

func TestValidateSkin(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/skins/cool.yaml",
		},
		"toast": {
			f:   "testdata/skins/toast.yaml",
			err: `Additional property bodys is not allowed`,
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.SkinSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestValidateK9s(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/k9s/cool.yaml",
		},
		"toast": {
			f:   "testdata/k9s/toast.yaml",
			err: `Additional property shellPods is not allowed`,
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.K9sSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestValidateContext(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/context/cool.yaml",
		},
		"toast": {
			f: "testdata/context/toast.yaml",
			err: `Additional property fred is not allowed
Additional property namespaces is not allowed`,
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.ContextSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestValidateAliases(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/aliases/cool.yaml",
		},
		"toast": {
			f: "testdata/aliases/toast.yaml",
			err: `Additional property alias is not allowed
aliases is required`,
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.AliasesSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}

func TestValidateViews(t *testing.T) {
	uu := map[string]struct {
		f   string
		err string
	}{
		"happy": {
			f: "testdata/views/cool.yaml",
		},
		"toast": {
			f: "testdata/views/toast.yaml",
			err: `Additional property cols is not allowed
Additional property sortCol is not allowed
Invalid type. Expected: object, given: null
columns is required`,
		},
	}

	v := json.NewValidator()
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			bb, err := os.ReadFile(u.f)
			assert.NoError(t, err)
			if err := v.Validate(json.ViewsSchema, bb); err != nil {
				assert.Equal(t, u.err, err.Error())
			}
		})
	}
}
