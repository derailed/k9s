// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func init() {
	dao.MetaAccess.RegisterMeta(client.NodeAllocGVR.String(), &metav1.APIResource{
		Name:         "nodes-alloc",
		SingularName: "node-alloc",
		Namespaced:   false,
		Kind:         "NodeAlloc",
		Verbs:        []string{"get", "list", "watch"},
		Categories:   []string{"k9s"},
	})
}

func TestNodeAllocNew(t *testing.T) {
	na := view.NewNodeAlloc(client.NodeAllocGVR)

	require.NoError(t, na.Init(makeCtx(t)))
	assert.Equal(t, "NodeAlloc", na.Name())
	assert.Greater(t, len(na.Hints()), 0)
}

func TestNodeAllocContext(t *testing.T) {
	na := view.NewNodeAlloc(client.NodeAllocGVR)
	require.NoError(t, na.Init(makeCtx(t)))

	// Verify the view is properly initialized
	assert.Equal(t, "NodeAlloc", na.Name())
}
