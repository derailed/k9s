// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Namespace)(nil)

// Namespace represents a namespace resource.
type Namespace struct {
	Resource
}

// mockNamespace returns static namespaces (fake DAO) for offline mode.
type mockNamespace struct {
	Namespace
}

// List always returns static namespaces.
func (n *mockNamespace) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	// Static namespaces.
	return []runtime.Object{
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name":              "default",
					"creationTimestamp": "2020-01-01T00:00:00Z",
				},
				"status": map[string]interface{}{
					"phase": "Active",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Namespace",
				"metadata": map[string]interface{}{
					"name":              "kube-system",
					"creationTimestamp": "2020-01-01T00:00:00Z",
				},
				"status": map[string]interface{}{
					"phase": "Active",
				},
			},
		},
	}, nil
}

// Get returns a static namespace.
func (n *mockNamespace) Get(ctx context.Context, path string) (runtime.Object, error) {
	oo, _ := n.List(ctx, "")
	if len(oo) > 0 {
		return oo[0], nil
	}
	return nil, fmt.Errorf("no static namespace")
}
