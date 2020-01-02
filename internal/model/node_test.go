package model_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"testing"

	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestNodeHydrate(t *testing.T) {
	f := makeFactory()
	var no model.Node
	no.Init("", "v1/nodes", f)

	o := render.NodeWithMetrics{Raw: load(t, "n1")}
	rr := make(render.Rows, 1)
	assert.Nil(t, no.Hydrate([]runtime.Object{&o}, rr, render.Node{}))
	assert.Equal(t, 1, len(rr))
	assert.Equal(t, "minikube", rr[0].ID)
	assert.Equal(t, render.Fields{
		"minikube",
		"Ready",
		"master",
		"v1.17.0",
		"4.19.81",
		"192.168.64.6",
		"<none>",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
		"n/a",
	}, rr[0].Fields[:len(rr[0].Fields)-1])
}

func BenchmarkNodeHydrate(b *testing.B) {
	f := makeFactory()
	var no model.Node
	no.Init("", "v1/nodes", f)
	o := load(b, "n1")
	rr := make(render.Rows, 1)

	oo := []runtime.Object{&render.NodeWithMetrics{Raw: o}}
	re := render.Node{}
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		no.Hydrate(oo, rr, re)
	}
}

// Helpers...

func load(t assert.TestingT, n string) *unstructured.Unstructured {
	raw, err := ioutil.ReadFile(fmt.Sprintf("test_assets/%s.json", n))
	assert.Nil(t, err)

	var o unstructured.Unstructured
	err = json.Unmarshal(raw, &o)
	assert.Nil(t, err)

	return &o
}
