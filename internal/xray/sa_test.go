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

func TestSARender(t *testing.T) {
	uu := map[string]struct {
		file           string
		level1, level2 int
		status         string
	}{
		"plain": {
			file:   "sa",
			level1: 1,
			level2: 2,
			status: xray.OkStatus,
		},
	}

	var re xray.ServiceAccount
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			o := load(t, u.file)
			root := xray.NewTreeNode("serviceaccounts", "serviceaccounts")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", o))
			assert.Equal(t, u.level1, root.CountChildren())
			assert.Equal(t, u.level2, root.Children[0].CountChildren())
		})
	}
}
