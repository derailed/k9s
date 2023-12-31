// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package xray_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/xray"
	"github.com/stretchr/testify/assert"
	metav1beta1 "k8s.io/apimachinery/pkg/apis/meta/v1beta1"
)

func TestGenericRender(t *testing.T) {
	uu := map[string]struct {
		level1 int
		status string
	}{
		"plain": {
			level1: 1,
			status: xray.OkStatus,
		},
	}

	var re xray.Generic
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			root := xray.NewTreeNode("generics", "generics")
			ctx := context.WithValue(context.Background(), xray.KeyParent, root)
			ctx = context.WithValue(ctx, internal.KeyFactory, makeFactory())

			assert.Nil(t, re.Render(ctx, "", makeTable()))
			assert.Equal(t, u.level1, root.CountChildren())
		})
	}
}

// Helpers...

func makeTable() metav1beta1.TableRow {
	return metav1beta1.TableRow{
		Cells: []interface{}{"fred", "blee"},
	}
}
