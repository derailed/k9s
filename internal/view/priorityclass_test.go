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

func TestPriorityClassNew(t *testing.T) {
	s := view.NewPriorityClass(client.PcGVR)

	require.NoError(t, s.Init(makeCtx()))
	assert.Equal(t, "PriorityClass", s.Name())
	assert.Len(t, s.Hints(), 6)
}
