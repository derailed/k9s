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
	"github.com/derailed/k9s/internal/view/cmd"
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

func TestAliasResolve(t *testing.T) {
	uu := map[string]struct {
		exp string
		ok  bool
		gvr *client.GVR
		cmd *cmd.Interpreter
	}{
		"gvr": {
			exp: "v1/pods",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods"),
		},

		"kind": {
			exp: "pod",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods"),
		},

		"plural": {
			exp: "pods",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods"),
		},

		"short-name": {
			exp: "po",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods"),
		},

		"short-name-with-args": {
			exp: "po 'a in (b,c)' @zorb bozo",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods 'a in (b,c)' @zorb bozo"),
		},

		"alias": {
			exp: "pipo",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods"),
		},

		"toast-command": {
			exp: "zorg",
		},

		"alias-no-args": {
			exp: "wkl",
			ok:  true,
			gvr: client.WkGVR,
			cmd: cmd.NewInterpreter("workloads"),
		},

		"alias-ns-arg": {
			exp: "pp",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods default"),
		},

		"multi-alias-ns-inception": {
			exp: "ppo",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods 'a=b,b=c' default"),
		},

		"full-alias": {
			exp: "ppc",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods @fred 'app=fred' default"),
		},

		"plain-filter": {
			exp: "po /fred @bozo ns-1",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods /fred @bozo ns-1"),
		},

		"alias-filter": {
			exp: "pipo /fred @bozo ns-1",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods /fred @bozo ns-1"),
		},

		"complex-filter": {
			exp: "ppc /fred @bozo ns-1",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods @bozo /fred 'app=fred' ns-1"),
		},

		"filtered": {
			exp: "pc",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods /cilium kube-system"),
		},

		"labels-in": {
			exp: "ppp",
			ok:  true,
			gvr: client.PodGVR,
			cmd: cmd.NewInterpreter("v1/pods 'app in (be,fe)'"),
		},
	}

	a := config.NewAliases()
	a.Define(client.PodGVR, "po", "pipo", "pod")
	a.Define(client.PodGVR, client.PodGVR.String())
	a.Define(client.PodGVR, client.PodGVR.AsResourceName())
	a.Define(client.WkGVR, client.WkGVR.String(), "workload", "wkl")
	a.Define(client.NewGVR("pod default"), "pp")
	a.Define(client.NewGVR("pipo a=b,b=c default"), "ppo")
	a.Define(client.NewGVR("pod default app=fred @fred"), "ppc")
	a.Define(client.NewGVR("pod /cilium kube-system"), "pc")
	a.Define(client.NewGVR("pod 'app in (be,fe)'"), "ppp")
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			p := cmd.NewInterpreter(u.exp)
			gvr, ok := a.Resolve(p)
			assert.Equal(t, u.ok, ok)
			if ok {
				assert.Equal(t, u.gvr, gvr)
				assert.Equal(t, u.cmd.GetLine(), p.GetLine())
			}
		})
	}
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
