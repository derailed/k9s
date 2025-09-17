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
