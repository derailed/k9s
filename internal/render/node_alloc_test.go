// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNodeAllocRender(t *testing.T) {
	pom := render.NodeWithMetrics{
		Raw:             load(t, "no"),
		MX:              makeNodeMX("n1", "10m", "20Mi"),
		PodCount:        5,
		RequestedCPU:    5000,
		RequestedMemory: 104857600,
	}

	var no render.NodeAlloc
	r := model1.NewRow(14)
	err := no.Render(&pom, "", &r)
	require.NoError(t, err)

	assert.Equal(t, "minikube", r.ID)
	assert.Equal(t, "minikube", r.Fields[0])
	assert.Equal(t, "5", r.Fields[7])
	assert.Contains(t, r.Fields, "5")
}

func TestNodeAllocRenderWithNAValues(t *testing.T) {
	pom := render.NodeWithMetrics{
		Raw:             load(t, "no"),
		MX:              makeNodeMX("n1", "10m", "20Mi"),
		PodCount:        -1,
		RequestedCPU:    -1,
		RequestedMemory: -1,
	}

	var no render.NodeAlloc
	r := model1.NewRow(14)
	err := no.Render(&pom, "", &r)
	require.NoError(t, err)

	assert.Equal(t, "minikube", r.ID)
	assert.Contains(t, r.Fields, "n/a")
}

func TestNodeAllocHeader(t *testing.T) {
	var no render.NodeAlloc
	h := no.Header("")

	assert.NotEmpty(t, h)
	assert.Equal(t, "NAME", h[0].Name)
	assert.Equal(t, "STATUS", h[1].Name)
	assert.Equal(t, "PODS", h[7].Name)
	assert.Equal(t, "CPU/R", h[10].Name)
	assert.Equal(t, "%CPU/R", h[12].Name)
	assert.Equal(t, "MEM/R", h[15].Name)
	assert.Equal(t, "%MEM/R", h[17].Name)
}

func BenchmarkNodeAllocRender(b *testing.B) {
	var (
		no  render.NodeAlloc
		r   = model1.NewRow(14)
		pom = render.NodeWithMetrics{
			Raw:             load(b, "no"),
			MX:              makeNodeMX("n1", "10m", "10Mi"),
			PodCount:        3,
			RequestedCPU:    3000,
			RequestedMemory: 31457280,
		}
	)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = no.Render(&pom, "", &r)
	}
}
