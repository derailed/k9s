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

func TestNSCleanser(t *testing.T) {
	ns := view.NewNamespace(client.NsGVR)

	require.NoError(t, ns.Init(makeCtx()))
	assert.Equal(t, "Namespaces", ns.Name())
	assert.Len(t, ns.Hints(), 7)
}
