// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestDaemonSetRender(t *testing.T) {
	uu := map[string]struct {
		file           string
		level1, level2 int
		status         string
	}{
		"plain": {
			file:   "ds",
			level1: 1,
			level2: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.DaemonSet
	for k := range uu {
		f := makeFactory()
		f.rows = map[*client.GVR][]runtime.Object{client.PodGVR: {load(t, "po")}}
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode(client.DsGVR, "daemonsets")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, f)

			require.NoError(t, re.Render(ctx, "", o))
			assert.Equal(t, u.level1, root.CountChildren())
			assert.Equal(t, u.level2, root.Children[0].CountChildren())
		})
	}
}
