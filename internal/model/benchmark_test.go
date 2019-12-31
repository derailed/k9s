package model_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestBenchmarkList(t *testing.T) {
	a := model.Benchmark{}
	a.Init(render.ClusterScope, "benchmarks", makeFactory())

	ctx := context.WithValue(context.Background(), internal.KeyDir, "test_assets/bench")
	oo, err := a.List(ctx)

	assert.Nil(t, err)
	assert.Equal(t, 1, len(oo))
	assert.Equal(t, "test_assets/bench/default_fred_1577308050814961000.txt", oo[0].(render.BenchInfo).Path)
}

func TestBenchmarkHydrate(t *testing.T) {
	a := model.Benchmark{}
	a.Init(render.ClusterScope, "benchmarks", makeFactory())

	ctx := context.WithValue(context.Background(), internal.KeyDir, "test_assets/bench")
	oo, err := a.List(ctx)
	assert.Nil(t, err)

	rr := make(render.Rows, len(oo))
	assert.Nil(t, a.Hydrate(oo, rr, render.Benchmark{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, "test_assets/bench/default_fred_1577308050814961000.txt", rr[0].ID)
	assert.Equal(t, render.Fields{
		"default",
		"fred",
		"fail",
		"816.6403",
		"0.0122",
		"0",
		"0",
		"default_fred_1577308050814961000.txt",
	},
		rr[0].Fields[:len(rr[0].Fields)-1],
	)
}
