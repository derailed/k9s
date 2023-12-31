// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
)

func TestNamespaceRender(t *testing.T) {
	uu := map[string]struct {
		file   string
		level1 int
		status string
	}{
		"plain": {
			file:   "ns",
			level1: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.Namespace
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode("namespaces", "namespaces")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", o))
			assert.Equal(t, u.level1, root.CountChildren())
		})
	}
}
