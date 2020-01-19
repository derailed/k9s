package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
)

func TestPodRender(t *testing.T) {
	uu := map[string]struct {
		file           string
		level1, level2 int
		status         string
	}{
		"plain": {
			file:   "po",
			level1: 1,
			level2: 2,
			status: xray.OkStatus,
		},
		"withInit": {
			file:   "init",
			level1: 1,
			level2: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.Pod
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode("pods", "pods")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", &render.PodWithMetrics{Raw: o}))
			assert.Equal(t, u.level1, root.CountChildren())
			assert.Equal(t, u.level2, root.Children[0].CountChildren())
		})
	}
}
