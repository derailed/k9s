// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/runtime"
)

func TestStatefulSetRender(t *testing.T) {
	uu := map[string]struct {
		file           string
		level1, level2 int
		status         string
	}{
		"plain": {
			file:   "sts",
			level1: 1,
			level2: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.StatefulSet
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			f := makeFactory()
			f.rows = map[string][]runtime.Object{"v1/pods": {load(t, "po")}}

			o := load(t, u.file)
			root := xray.NewTreeNode("statefulsets", "statefulsets")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, f)

			assert.Nil(t, re.Render(ctx, "", o))
			assert.Equal(t, u.level1, root.CountChildren())
			assert.Equal(t, u.level2, root.Children[0].CountChildren())
		})
	}
}
