// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"fmt"
	"os"
	"path"
	"slices"
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/stretchr/testify/assert"
)

func TestAliasClear(t *testing.T) {
	a := testAliases()
	a.Clear()

	assert.Equal(t, 0, len(a.Keys()))
}

func TestAliasKeys(t *testing.T) {
	a := testAliases()
	kk := a.Keys()
	slices.Sort(kk)

	assert.Equal(t, []string{"a1", "a11", "a2", "a3"}, kk)
}

func TestAliasShortNames(t *testing.T) {
	a := testAliases()
	ess := config.ShortNames{
		"gvr1": []string{"a1", "a11"},
		"gvr2": []string{"a2"},
		"gvr3": []string{"a3"},
	}
	ss := a.ShortNames()
	assert.Equal(t, len(ess), len(ss))
	for k, v := range ss {
		v1, ok := ess[k]
		assert.True(t, ok, fmt.Sprintf("missing: %q", k))
		slices.Sort(v)
		assert.Equal(t, v1, v)
	}
}

func TestAliasDefine(t *testing.T) {
	type aliasDef struct {
		cmd     string
		aliases []string
	}

	uu := map[string]struct {
		aliases            []aliasDef
		registeredCommands map[string]string
	}{
		"simple": {
			aliases: []aliasDef{
				{
					cmd:     "one",
					aliases: []string{"blee", "duh"},
				},
			},
			registeredCommands: map[string]string{
				"blee": "one",
				"duh":  "one",
			},
		},
		"duplicates": {
			aliases: []aliasDef{
				{
					cmd:     "one",
					aliases: []string{"blee", "duh"},
				}, {
					cmd:     "two",
					aliases: []string{"blee", "duh", "fred", "zorg"},
				},
			},
			registeredCommands: map[string]string{
				"blee": "one",
				"duh":  "one",
				"fred": "two",
				"zorg": "two",
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			configAlias := config.NewAliases()
			for _, aliases := range u.aliases {
				for _, a := range aliases.aliases {
					configAlias.Define(aliases.cmd, a)
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

	assert.Nil(t, a.Load(path.Join(config.AppConfigDir, "plain.yaml")))
	assert.Equal(t, 54, len(a.Alias))
}

func TestAliasesSave(t *testing.T) {
	assert.NoError(t, data.EnsureFullPath("/tmp/test-aliases", data.DefaultDirMod))
	defer assert.NoError(t, os.RemoveAll("/tmp/test-aliases"))

	config.AppAliasesFile = "/tmp/test-aliases/aliases.yaml"
	a := testAliases()
	c := len(a.Alias)

	assert.Equal(t, c, len(a.Alias))
	assert.Nil(t, a.Save())
	assert.Nil(t, a.LoadFile(config.AppAliasesFile))
	assert.Equal(t, c, len(a.Alias))
}

// Helpers...

func testAliases() *config.Aliases {
	a := config.NewAliases()
	a.Alias["a1"] = "gvr1"
	a.Alias["a11"] = "gvr1"
	a.Alias["a2"] = "gvr2"
	a.Alias["a3"] = "gvr3"

	return a
}
