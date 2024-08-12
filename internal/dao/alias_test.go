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
)

func TestAsGVR(t *testing.T) {
	a := dao.NewAlias(makeFactory())
	a.Aliases.Define("v1/pods", "po", "pod", "pods")
	a.Aliases.Define("workloads", "workloads", "workload", "wkl")

	uu := map[string]struct {
		cmd string
		ok  bool
		gvr client.GVR
	}{
		"ok": {
			cmd: "pods",
			ok:  true,
			gvr: client.NewGVR("v1/pods"),
		},
		"ok-short": {
			cmd: "po",
			ok:  true,
			gvr: client.NewGVR("v1/pods"),
		},
		"missing": {
			cmd: "zorg",
		},
		"alias": {
			cmd: "wkl",
			ok:  true,
			gvr: client.NewGVR("workloads"),
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			gvr, _, ok := a.AsGVR(u.cmd)
			assert.Equal(t, u.ok, ok)
			if u.ok {
				assert.Equal(t, u.gvr, gvr)
			}
		})
	}
}

func TestAliasList(t *testing.T) {
	a := dao.Alias{}
	a.Init(makeFactory(), client.NewGVR("aliases"))

	ctx := context.WithValue(context.Background(), internal.KeyAliases, makeAliases())
	oo, err := a.List(ctx, "-")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oo))
	assert.Equal(t, 2, len(oo[0].(render.AliasRes).Aliases))
}

// ----------------------------------------------------------------------------
// Helpers...

func makeAliases() *dao.Alias {
	return &dao.Alias{
		Aliases: &config.Aliases{
			Alias: config.Alias{
				"fred": "v1/fred",
				"f":    "v1/fred",
				"blee": "v1/blee",
				"b":    "v1/blee",
			},
		},
	}
}
