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

func TestDpRender(t *testing.T) {
	c := render.Deployment{}
	r := model1.NewRow(7)

	require.NoError(t, c.Render(load(t, "dp"), "", &r))
	assert.Equal(t, "icx/icx-db", r.ID)
	assert.Equal(t, model1.Fields{"icx", "icx-db", "n/a", "1/1", "1", "1"}, r.Fields[:6])
}

func BenchmarkDpRender(b *testing.B) {
	var (
		c = render.Deployment{}
		r = model1.NewRow(7)
		o = load(b, "dp")
	)

	b.ResetTimer()
	b.ReportAllocs()
	for range b.N {
		_ = c.Render(o, "", &r)
	}
}

func TestDpRenderFlagsMissingReplicas(t *testing.T) {
	uu := map[string]struct {
		replicas, statusReplicas, available int64
		valid                               bool
	}{
		"happy": {
			replicas: 1, statusReplicas: 1, available: 1, valid: true,
		},
		"no-pods-created": {
			replicas: 3, statusReplicas: 0, available: 0,
		},
		"some-pods-created": {
			replicas: 3, statusReplicas: 1, available: 1,
		},
		"pod-not-available": {
			replicas: 1, statusReplicas: 1, available: 0,
		},
		"scaled-to-zero": {
			replicas: 0, statusReplicas: 0, available: 0, valid: true,
		},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			o := load(t, "dp")
			o.Object["spec"].(map[string]any)["replicas"] = u.replicas
			st := o.Object["status"].(map[string]any)
			st["replicas"] = u.statusReplicas
			st["availableReplicas"] = u.available

			c := render.Deployment{}
			r := model1.NewRow(9)
			require.NoError(t, c.Render(o, "", &r))
			assert.Equal(t, u.valid, r.Fields[7] == "", "valid column %q", r.Fields[7])
		})
	}
}
