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

func TestReferenceNew(t *testing.T) {
	s := view.NewReference(client.RefGVR)

	require.NoError(t, s.Init(makeCtx(t)))
	assert.Equal(t, "References", s.Name())
	assert.Len(t, s.Hints(), 4)
}
