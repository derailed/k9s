// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/stretchr/testify/assert"
)

func TestNewDir(t *testing.T) {
	d := dao.NewDir(nil)
	ctx := context.WithValue(context.Background(), internal.KeyPath, "testdata/dir")
	oo, err := d.List(ctx, "")

	assert.Nil(t, err)
	assert.Equal(t, 2, len(oo))
}
