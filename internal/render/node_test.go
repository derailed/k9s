// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeRender(t *testing.T) {
	pom := render.NodeWithMetrics{
		Raw: load(t, "no"),
		MX:  makeNodeMX("n1", "10m", "20Mi"),
	}

	var no render.Node
	r := model1.NewRow(14)
	err := no.Render(&pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "minikube", r.ID)
	e := model1.Fields{
		"minikube",            // NAME
		"Ready",               // STATUS
		"master",              // ROLE
		"amd64",               // ARCH
		"0",                   // TAINTS
		"",                    // TAINTS-LIST
		"v1.15.2",             // VERSION
		"Buildroot 2018.05.3", // OS-IMAGE
		"4.15.0",              // KERNEL
		"192.168.64.107",      // INTERNAL-IP
		"<none>",              // EXTERNAL-IP
		"0",                   // PODS
		"10",                  // CPU
		"20",                  // MEM
		"0",                   // %CPU
		"0",                   // %MEM
		"4000",                // CPU/A
		"7874",                // MEM/A
	}
	assert.Equal(t, e, r.Fields[:18])
}

func BenchmarkNodeRender(b *testing.B) {
	pom := render.NodeWithMetrics{
		Raw: load(b, "no"),
		MX:  makeNodeMX("n1", "10m", "10Mi"),
	}
	var no render.Node
	r := model1.NewRow(14)
	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = no.Render(&pom, "", &r)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeNodeMX(name, cpu, mem string) *mv1beta1.NodeMetrics {
	return &mv1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Usage: makeRes(cpu, mem),
	}
}
