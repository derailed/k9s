// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNamespaceFromSelector(t *testing.T) {
	uu := map[string]struct {
		ownNS      string
		anyNS      bool
		matchNames []string
		want       string
	}{
		"any wins":               {ownNS: "monitoring", anyNS: true, matchNames: []string{"foo"}, want: client.NamespaceAll},
		"single matchName":       {ownNS: "monitoring", matchNames: []string{"foo"}, want: "foo"},
		"multiple matchNames":    {ownNS: "monitoring", matchNames: []string{"foo", "bar"}, want: client.NamespaceAll},
		"empty falls back to own": {ownNS: "monitoring", want: "monitoring"},
	}

	for name, u := range uu {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, u.want, namespaceFromSelector(u.ownNS, u.anyNS, u.matchNames))
		})
	}
}

func TestTargetFromSpec(t *testing.T) {
	uu := map[string]struct {
		obj          map[string]any
		wantNS       string
		wantSelector string
		wantEmpty    bool
		wantErr      bool
	}{
		"matchLabels only": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec": map[string]any{
					"selector": map[string]any{
						"matchLabels": map[string]any{"app": "api"},
					},
				},
			},
			wantNS:       "monitoring",
			wantSelector: "app=api",
		},
		"matchExpressions": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec": map[string]any{
					"selector": map[string]any{
						"matchExpressions": []any{
							map[string]any{
								"key":      "tier",
								"operator": "In",
								"values":   []any{"backend", "frontend"},
							},
						},
					},
				},
			},
			wantNS:       "monitoring",
			wantSelector: "tier in (backend,frontend)",
		},
		"namespaceSelector any": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec": map[string]any{
					"selector":          map[string]any{"matchLabels": map[string]any{"app": "api"}},
					"namespaceSelector": map[string]any{"any": true},
				},
			},
			wantNS:       client.NamespaceAll,
			wantSelector: "app=api",
		},
		"namespaceSelector matchNames single": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec": map[string]any{
					"selector":          map[string]any{"matchLabels": map[string]any{"app": "api"}},
					"namespaceSelector": map[string]any{"matchNames": []any{"prod"}},
				},
			},
			wantNS:       "prod",
			wantSelector: "app=api",
		},
		"namespaceSelector matchNames multi widens to all": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec": map[string]any{
					"selector":          map[string]any{"matchLabels": map[string]any{"app": "api"}},
					"namespaceSelector": map[string]any{"matchNames": []any{"prod", "staging"}},
				},
			},
			wantNS:       client.NamespaceAll,
			wantSelector: "app=api",
		},
		"missing selector is empty": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec":     map[string]any{},
			},
			wantNS:    "monitoring",
			wantEmpty: true,
		},
		"invalid selector type errors": {
			obj: map[string]any{
				"metadata": map[string]any{"namespace": "monitoring"},
				"spec":     map[string]any{"selector": "not-a-map"},
			},
			wantErr: true,
		},
	}

	for name, u := range uu {
		t.Run(name, func(t *testing.T) {
			un := &unstructured.Unstructured{Object: u.obj}
			got, err := targetFromSpec(un, "spec")
			if u.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, u.wantNS, got.Namespace)
			if u.wantEmpty {
				assert.True(t, got.Selector.Empty())
			} else {
				assert.Equal(t, u.wantSelector, got.Selector.String())
			}
		})
	}
}

func TestProbeTargetFromSpec(t *testing.T) {
	obj := map[string]any{
		"metadata": map[string]any{"namespace": "monitoring"},
		"spec": map[string]any{
			"targets": map[string]any{
				"ingress": map[string]any{
					"selector":          map[string]any{"matchLabels": map[string]any{"tier": "edge"}},
					"namespaceSelector": map[string]any{"matchNames": []any{"prod"}},
				},
			},
		},
	}
	un := &unstructured.Unstructured{Object: obj}
	got, err := targetFromSpec(un, "spec", "targets", "ingress")
	require.NoError(t, err)
	assert.Equal(t, "prod", got.Namespace)
	assert.Equal(t, "tier=edge", got.Selector.String())
}
