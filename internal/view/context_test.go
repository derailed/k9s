// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestContext(t *testing.T) {
	ctx := view.NewContext(client.CtGVR)

	require.NoError(t, ctx.Init(makeCtx()))
	assert.Equal(t, "Contexts", ctx.Name())
	assert.Len(t, ctx.Hints(), 5)
}
