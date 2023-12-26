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
)

func TestBenchmarkList(t *testing.T) {
	a := dao.Benchmark{}
	a.Init(makeFactory(), client.NewGVR("benchmarks"))

	ctx := context.WithValue(context.Background(), internal.KeyDir, "testdata/bench")
	ctx = context.WithValue(ctx, internal.KeyPath, "")
	oo, err := a.List(ctx, "-")

	assert.Nil(t, err)
	assert.Equal(t, 1, len(oo))
	assert.Equal(t, "testdata/bench/default_fred_1577308050814961000.txt", oo[0].(render.BenchInfo).Path)
}
