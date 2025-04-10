// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodRender(t *testing.T) {
	uu := map[string]struct {
		file            string
		count, children int
		status          string
	}{
		"plain": {
			file:     "po",
			children: 1,
			count:    7,
			status:   xray.OkStatus,
		},
		"withInit": {
			file:     "init",
			children: 1,
			count:    7,
			status:   xray.OkStatus,
		},
		"cilium": {
			file:     "cilium",
			children: 1,
			count:    8,
			status:   xray.OkStatus,
		},
	}

	var re xray.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode(client.PodGVR, "pods")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			require.NoError(t, re.Render(ctx, "", &render.PodWithMetrics{Raw: o}))
			assert.Equal(t, u.children, root.CountChildren())
			assert.Equal(t, u.count, root.Count(client.NoGVR))
		})
	}
}
