package model_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestPodHydrate(t *testing.T) {
	f := makeFactory()
	var po model.Pod
	po.Init("", "v1/pods", f)

	o := render.PodWithMetrics{Raw: load(t, "p1")}
	rr := make(render.Rows, 1)
	assert.Nil(t, po.Hydrate([]runtime.Object{&o}, rr, render.Pod{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, "default/nginx-7fb78fb6d8-2w75j", rr[0].ID)
	assert.Equal(t, render.Fields{
		"default",
		"nginx-7fb78fb6d8-2w75j",
		"1/1",
		"Running",
		"0",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
		"10.44.0.229",
		"gke-k9s-default-pool-0fa2fb89-lbtf",
		"GA",
	}, rr[0].Fields[:len(rr[0].Fields)-1])
}

func BenchmarkPodHydrate(b *testing.B) {
	f := makeFactory()
	var po model.Pod
	po.Init("", "v1/pods", f)
	o := load(b, "p1")
	rr := make(render.Rows, 1)

	oo := []runtime.Object{&render.PodWithMetrics{Raw: o}}
	re := render.Pod{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		po.Hydrate(oo, rr, re)
	}
}
