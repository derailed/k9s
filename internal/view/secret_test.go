// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/mock"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSecretNew(t *testing.T) {
	s := view.NewSecret(client.SecGVR)

	require.NoError(t, s.Init(makeCtx(t)))
	assert.Equal(t, "Secrets", s.Name())
	assert.Len(t, s.Hints(), 12)
	assert.True(t, hintHas(s.Hints(), "Edit Decoded"))
	assert.True(t, hintHas(s.Hints(), "Edit Raw"))
}

func TestSecretEditKeysReadOnly(t *testing.T) {
	cfg := mock.NewMockConfig(t)
	cfg.K9s.ReadOnly = true
	ctx := context.WithValue(context.Background(), internal.KeyApp, view.NewApp(cfg))
	s := view.NewSecret(client.SecGVR)

	require.NoError(t, s.Init(ctx))
	assert.False(t, hintHas(s.Hints(), "Edit Decoded"))
	assert.False(t, hintHas(s.Hints(), "Edit Raw"))
}

func hintHas(hints model.MenuHints, desc string) bool {
	for _, h := range hints {
		if h.Description == desc {
			return true
		}
	}
	return false
}
