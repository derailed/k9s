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

func TestReplicaSetRender(t *testing.T) {
	c := render.ReplicaSet{}
	r := model1.NewRow(4)

	require.NoError(t, c.Render(load(t, "rs"), "", &r))
	assert.Equal(t, "icx/icx-db-7d4b578979", r.ID)
	assert.Equal(t, model1.Fields{"icx", "icx-db-7d4b578979", "n/a", "1", "1", "1"}, r.Fields[:6])
}

func TestReplicaSetRenderValid(t *testing.T) {
	uu := map[string]struct {
		replicas, statusReplicas, ready int64
		valid                           string
	}{
		"happy":             {replicas: 1, statusReplicas: 1, ready: 1},
		"no-pods-created":   {replicas: 3, statusReplicas: 0, ready: 0, valid: "want 3 replicas got 0 ready"},
		"some-pods-created": {replicas: 3, statusReplicas: 1, ready: 1, valid: "want 3 replicas got 1 ready"},
		"pod-not-ready":     {replicas: 1, statusReplicas: 1, ready: 0, valid: "mismatch desired(1) vs ready(0)"},
		"scaled-to-zero":    {replicas: 0, statusReplicas: 0, ready: 0},
	}

	for k, u := range uu {
		t.Run(k, func(t *testing.T) {
			o := load(t, "rs")
			o.Object["spec"].(map[string]any)["replicas"] = u.replicas
			st := o.Object["status"].(map[string]any)
			st["replicas"] = u.statusReplicas
			st["readyReplicas"] = u.ready

			c := render.ReplicaSet{}
			r := model1.NewRow(11)
			require.NoError(t, c.Render(o, "", &r))
			assert.Equal(t, u.valid, r.Fields[9])
		})
	}
}
