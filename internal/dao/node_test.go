// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func makeNodeFactory(nodes, podsPerNode int) dao.Factory {
	nn := make([]runtime.Object, 0, nodes)
	pp := make([]runtime.Object, 0, nodes*podsPerNode)
	for n := range nodes {
		name := fmt.Sprintf("node-%d", n)
		nn = append(nn, &unstructured.Unstructured{Object: map[string]any{
			"metadata": map[string]any{"name": name},
		}})
		for p := range podsPerNode {
			pp = append(pp, &unstructured.Unstructured{Object: map[string]any{
				"metadata": map[string]any{"namespace": "default", "name": fmt.Sprintf("%s-pod-%d", name, p)},
				"spec":     map[string]any{"nodeName": name},
			}})
		}
	}

	return &testFactory{inventory: map[string]map[*client.GVR][]runtime.Object{
		client.BlankNamespace: {client.NodeGVR: nn, client.PodGVR: pp},
	}}
}

func nodeListContext() context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyPodCounting, true)
	return context.WithValue(ctx, internal.KeyWithMetrics, false)
}

func TestNodeListPodCount(t *testing.T) {
	f := makeNodeFactory(3, 2)
	oo, err := f.List(client.PodGVR, client.BlankNamespace, false, nil)
	require.NoError(t, err)
	// A pod that is not scheduled yet belongs to no node.
	oo = append(oo, &unstructured.Unstructured{Object: map[string]any{
		"metadata": map[string]any{"namespace": "default", "name": "pending"},
		"spec":     map[string]any{},
	}})
	f.(*testFactory).inventory[client.BlankNamespace][client.PodGVR] = oo

	var n dao.Node
	n.Init(f, client.NodeGVR)
	nn, err := n.List(nodeListContext(), client.BlankNamespace)
	require.NoError(t, err)
	require.Len(t, nn, 3)
	for _, o := range nn {
		nmx, ok := o.(*render.NodeWithMetrics)
		require.True(t, ok)
		assert.Equal(t, 2, nmx.PodCount, nmx.Raw.GetName())
	}
}

func TestNodeGetPodCount(t *testing.T) {
	var n dao.Node
	n.Init(makeNodeFactory(3, 2), client.NodeGVR)

	o, err := n.Get(nodeListContext(), "node-1")
	require.NoError(t, err)
	nmx, ok := o.(*render.NodeWithMetrics)
	require.True(t, ok)
	assert.Equal(t, 2, nmx.PodCount)
}

func BenchmarkNodeList(b *testing.B) {
	for _, nodes := range []int{50, 200, 1_000} {
		const podsPerNode = 30
		b.Run(fmt.Sprintf("nodes=%d/pods=%d", nodes, nodes*podsPerNode), func(b *testing.B) {
			var n dao.Node
			n.Init(makeNodeFactory(nodes, podsPerNode), client.NodeGVR)
			ctx := nodeListContext()

			b.ReportAllocs()
			b.ResetTimer()
			for range b.N {
				if _, err := n.List(ctx, client.BlankNamespace); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkNodeGet(b *testing.B) {
	const nodes, podsPerNode = 200, 30
	var n dao.Node
	n.Init(makeNodeFactory(nodes, podsPerNode), client.NodeGVR)
	ctx := nodeListContext()

	b.ReportAllocs()
	b.ResetTimer()
	for range b.N {
		if _, err := n.Get(ctx, "node-1"); err != nil {
			b.Fatal(err)
		}
	}
}
