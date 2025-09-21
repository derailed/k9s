// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

func podOnNode(nodeSpec any) runtime.Object {
	u := &unstructured.Unstructured{
		Object: map[string]any{
			"apiVersion": "v1",
			"kind":       "Pod",
			"spec":       map[string]any{},
		},
	}
	if nodeSpec != nil {
		u.Object["spec"].(map[string]any)["nodeName"] = nodeSpec
	}
	return u
}

func TestCountPodsByNode(t *testing.T) {
	n := &Node{}

	tests := []struct {
		name    string
		pods    []runtime.Object
		want    map[string]int
		wantErr bool
	}{
		{
			name: "no pods",
			pods: []runtime.Object{},
			want: map[string]int{},
		},
		{
			name: "single pod on node-a",
			pods: []runtime.Object{
				podOnNode("node-a"),
			},
			want: map[string]int{"node-a": 1},
		},
		{
			name: "multiple pods across nodes",
			pods: []runtime.Object{
				podOnNode("node-a"),
				podOnNode("node-a"),
				podOnNode("node-b"),
			},
			want: map[string]int{"node-a": 2, "node-b": 1},
		},
		{
			name: "missing nodeName key (skip)",
			pods: []runtime.Object{
				podOnNode(nil),
			},
			want: map[string]int{},
		},
		{
			name: "empty nodeName string (skip)",
			pods: []runtime.Object{
				podOnNode(""),
			},
			want: map[string]int{},
		},
		{
			name: "unexpected nodeName type (skip)",
			pods: []runtime.Object{
				podOnNode(123),
				podOnNode(true),
				podOnNode(map[string]any{"x": "y"}),
			},
			want: map[string]int{},
		},
		{
			name: "spec is not a map (err)",
			pods: []runtime.Object{
				&unstructured.Unstructured{
					Object: map[string]any{
						"apiVersion": "v1",
						"kind":       "Pod",
						"spec":       "oh no!",
					},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := n.CountPodsByNode(tt.pods)

			if tt.wantErr {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
