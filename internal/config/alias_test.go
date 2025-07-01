// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"maps"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAliasClear(t *testing.T) {
	a := testAliases()
	a.Clear()

	assert.Empty(t, slices.Collect(maps.Keys(a.Alias)))
}

func TestAliasKeys(t *testing.T) {
	a := testAliases()
	kk := maps.Keys(a.Alias)

	assert.Equal(t, []string{"a1", "a11", "a2", "a3"}, slices.Sorted(kk))
}

func TestAliasShortNames(t *testing.T) {
	a := testAliases()
	ess := config.ShortNames{
		gvr1: []string{"a1", "a11"},
		gvr2: []string{"a2"},
		gvr3: []string{"a3"},
	}
	ss := a.ShortNames()
	assert.Len(t, ss, len(ess))
	for k, v := range ss {
		v1, ok := ess[k]
		assert.True(t, ok, "missing: %q", k)
		slices.Sort(v)
		assert.Equal(t, v1, v)
	}
}

func TestAliasDefine(t *testing.T) {
	type aliasDef struct {
		gvr     *client.GVR
		aliases []string
	}

	uu := map[string]struct {
		aliases            []aliasDef
		registeredCommands map[string]*client.GVR
	}{
		"simple": {
			aliases: []aliasDef{
				{
					gvr:     client.NewGVR("one"),
					aliases: []string{"blee", "duh"},
				},
			},
			registeredCommands: map[string]*client.GVR{
				"blee": client.NewGVR("one"),
				"duh":  client.NewGVR("one"),
			},
		},
		"duplicates": {
			aliases: []aliasDef{
				{
					gvr:     client.NewGVR("one"),
					aliases: []string{"blee", "duh"},
				}, {
					gvr:     client.NewGVR("two"),
					aliases: []string{"blee", "duh", "fred", "zorg"},
				},
			},
			registeredCommands: map[string]*client.GVR{
				"blee": client.NewGVR("one"),
				"duh":  client.NewGVR("one"),
				"fred": client.NewGVR("two"),
				"zorg": client.NewGVR("two"),
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			configAlias := config.NewAliases()
			for _, aliases := range u.aliases {
				for _, a := range aliases.aliases {
					configAlias.Define(aliases.gvr, a)
				}
			}
			for alias, cmd := range u.registeredCommands {
				v, ok := configAlias.Get(alias)
				assert.True(t, ok)
				assert.Equal(t, cmd, v, "Wrong command for alias "+alias)
			}
		})
	}
}

func TestAliasesLoad(t *testing.T) {
	config.AppConfigDir = "testdata/aliases"
	a := config.NewAliases()
	require.NoError(t, a.Load(path.Join(config.AppConfigDir, "plain.yaml")))

	assert.Len(t, a.Alias, 55)
}

func TestAliasesSave(t *testing.T) {
	require.NoError(t, data.EnsureFullPath("/tmp/test-aliases", data.DefaultDirMod))
	defer require.NoError(t, os.RemoveAll("/tmp/test-aliases"))

	config.AppAliasesFile = "/tmp/test-aliases/aliases.yaml"
	a := testAliases()
	c := len(a.Alias)

	assert.Len(t, a.Alias, c)
	require.NoError(t, a.Save())
	require.NoError(t, a.LoadFile(config.AppAliasesFile))
	assert.Len(t, a.Alias, c)
}

// Helpers...

var (
	gvr1 = client.NewGVR("gvr1")
	gvr2 = client.NewGVR("gvr2")
	gvr3 = client.NewGVR("gvr3")
)

func testAliases() *config.Aliases {
	a := config.NewAliases()
	a.Alias["a1"] = gvr1
	a.Alias["a11"] = gvr1
	a.Alias["a2"] = gvr2
	a.Alias["a3"] = gvr3

	return a
}
