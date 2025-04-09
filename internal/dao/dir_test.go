// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDir(t *testing.T) {
	d := dao.NewDir(nil)
	ctx := context.WithValue(context.Background(), internal.KeyPath, "testdata/dir")
	oo, err := d.List(ctx, "")

	require.NoError(t, err)
	assert.Len(t, oo, 2)
}
