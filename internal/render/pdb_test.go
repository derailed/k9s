// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPodDisruptionBudgetRender(t *testing.T) {
	c := render.PodDisruptionBudget{}
	r := model1.NewRow(9)

	require.NoError(t, c.Render(load(t, "pdb"), "", &r))
	assert.Equal(t, "default/fred", r.ID)
	assert.Equal(t, model1.Fields{"default", "fred", "2", render.NAValue, "0", "0", "2", "0"}, r.Fields[:8])
}

func TestPodDisruptionBudgetRenderValid(t *testing.T) {
	uu := map[string]struct {
		spec                           map[string]any
		currentHealthy, desiredHealthy int64
		valid                          string
	}{
		"percent-min-ok":  {spec: map[string]any{"minAvailable": "50%"}, currentHealthy: 2, desiredHealthy: 2},
		"percent-min-bad": {spec: map[string]any{"minAvailable": "50%"}, currentHealthy: 1, desiredHealthy: 2, valid: "expected 2 but got 1"},
		"int-min-ok":      {spec: map[string]any{"minAvailable": int64(2)}, currentHealthy: 2, desiredHealthy: 2},
		"int-min-bad":     {spec: map[string]any{"minAvailable": int64(2)}, currentHealthy: 1, desiredHealthy: 2, valid: "expected 2 but got 1"},
		"int-max-ok":      {spec: map[string]any{"maxUnavailable": int64(1)}, currentHealthy: 1, desiredHealthy: 1},
		"int-max-bad":     {spec: map[string]any{"maxUnavailable": int64(1)}, currentHealthy: 0, desiredHealthy: 1, valid: "expected 1 but got 0"},
		"percent-max-bad": {spec: map[string]any{"maxUnavailable": "50%"}, currentHealthy: 1, desiredHealthy: 2, valid: "expected 2 but got 1"},
		"no-budget":       {spec: map[string]any{}, currentHealthy: 0, desiredHealthy: 0},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			o := &unstructured.Unstructured{Object: map[string]any{
				"apiVersion": "policy/v1",
				"kind":       "PodDisruptionBudget",
				"metadata": map[string]any{
					"name":              "fred",
					"namespace":         "default",
					"creationTimestamp": "2026-01-01T00:00:00Z",
				},
				"spec": u.spec,
				"status": map[string]any{
					"currentHealthy":     u.currentHealthy,
					"desiredHealthy":     u.desiredHealthy,
					"expectedPods":       int64(4),
					"disruptionsAllowed": int64(0),
				},
			}}

			var c render.PodDisruptionBudget
			r := model1.NewRow(11)
			require.NoError(t, c.Render(o, "", &r))
			assert.Equal(t, u.valid, r.Fields[9])
		})
	}
}
