// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDir(t *testing.T) {
	v := view.NewDir("/fred")

	require.NoError(t, v.Init(makeCtx()))
	assert.Equal(t, "Directory", v.Name())
	assert.Len(t, v.Hints(), 7)
}
