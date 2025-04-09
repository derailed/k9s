// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBenchmarkList(t *testing.T) {
	a := dao.Benchmark{}
	a.Init(makeFactory(), client.BeGVR)

	ctx := context.WithValue(context.Background(), internal.KeyDir, "testdata/bench")
	ctx = context.WithValue(ctx, internal.KeyPath, "")
	oo, err := a.List(ctx, "-")

	require.NoError(t, err)
	assert.Len(t, oo, 1)
	assert.Equal(t, "testdata/bench/default_fred_1577308050814961000.txt", oo[0].(render.BenchInfo).Path)
}
