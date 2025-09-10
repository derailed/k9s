// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAsGVR(t *testing.T) {
	a := dao.NewAlias(makeFactory())
	a.Define(client.PodGVR, "po", "pipo", "pod")
	a.Define(client.PodGVR, client.PodGVR.String())
	a.Define(client.PodGVR, client.PodGVR.AsResourceName())
	a.Define(client.WkGVR, client.WkGVR.String(), "workload", "wkl")
	a.Define(client.NewGVR("pod default"), "pp")
	a.Define(client.NewGVR("pipo default"), "ppo")
	a.Define(client.NewGVR("pod default app=fred @fred"), "ppc")

	uu := map[string]struct {
		cmd string
		ok  bool
		gvr *client.GVR
		exp string
	}{
		"gvr": {
			cmd: "v1/pods",
			ok:  true,
			gvr: client.PodGVR,
		},

		"r": {
			cmd: "pods",
			ok:  true,
			gvr: client.PodGVR,
		},

		"alias1": {
			cmd: "po",
			ok:  true,
			gvr: client.PodGVR,
		},

		"alias-2": {
			cmd: "pipo",
			ok:  true,
			gvr: client.PodGVR,
		},

		"missing": {
			cmd: "zorg",
		},

		"no-args": {
			cmd: "wkl",
			ok:  true,
			gvr: client.WkGVR,
		},

		"ns-arg": {
			cmd: "pp",
			ok:  true,
			gvr: client.PodGVR,
			exp: "default",
		},

		"ns-inception": {
			cmd: "ppo",
			ok:  true,
			gvr: client.PodGVR,
			exp: "default",
		},

		"full-alias": {
			cmd: "ppc",
			ok:  true,
			gvr: client.PodGVR,
			exp: "default app=fred @fred",
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			gvr, exp, ok := a.AsGVR(u.cmd)
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.gvr, gvr)
				assert.Equal(t, u.exp, exp)
			}
		})
	}
}

func TestAliasList(t *testing.T) {
	a := dao.Alias{}
	a.Init(makeFactory(), client.AliGVR)

	ctx := context.WithValue(context.Background(), internal.KeyAliases, makeAliases())
	oo, err := a.List(ctx, "-")

	require.NoError(t, err)
	assert.Len(t, oo, 2)
	assert.Len(t, oo[0].(render.AliasRes).Aliases, 2)
}

// ----------------------------------------------------------------------------
// Helpers...

func makeAliases() *dao.Alias {
	gvr1 := client.NewGVR("v1/fred")
	gvr2 := client.NewGVR("v1/blee")

	return &dao.Alias{
		Aliases: &config.Aliases{
			Alias: config.Alias{
				"fred": gvr1,
				"f":    gvr1,
				"blee": gvr2,
				"b":    gvr2,
			},
		},
	}
}
