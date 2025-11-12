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

func TestContainerNew(t *testing.T) {
	c := view.NewContainer(client.CoGVR)

	require.NoError(t, c.Init(makeCtx(t)))
	assert.Equal(t, "Containers", c.Name())
	assert.Len(t, c.Hints(), 20)
}
