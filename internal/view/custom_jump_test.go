// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/stretchr/testify/assert"
)

// TestChainContext verifies context chaining preserves existing values and adds field selector.
// Regression test for issue #4139: custom jump was overwriting context instead of chaining it.
func TestChainContext(t *testing.T) {
	tests := []struct {
		name          string
		existingCtxFn ContextFunc
		fieldSelector string
		checkContext  func(t *testing.T, ctx context.Context)
	}{
		{
			name: "existing_and_field",
			existingCtxFn: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, internal.KeyPodCounting, true)
			},
			fieldSelector: "metadata.name=test",
			checkContext: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, true, ctx.Value(internal.KeyPodCounting))
				assert.Equal(t, "metadata.name=test", ctx.Value(internal.KeyFields))
			},
		},
		{
			name: "existing_only",
			existingCtxFn: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, internal.KeyPodCounting, false)
			},
			fieldSelector: "",
			checkContext: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, false, ctx.Value(internal.KeyPodCounting))
				assert.Nil(t, ctx.Value(internal.KeyFields))
			},
		},
		{
			name:          "field_only",
			existingCtxFn: nil,
			fieldSelector: "metadata.namespace=default",
			checkContext: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, "metadata.namespace=default", ctx.Value(internal.KeyFields))
			},
		},
		{
			name: "field_overriding_old",
			existingCtxFn: func(ctx context.Context) context.Context {
				return context.WithValue(ctx, internal.KeyFields, "old-selector")
			},
			fieldSelector: "new-selector",
			checkContext: func(t *testing.T, ctx context.Context) {
				assert.Equal(t, "new-selector", ctx.Value(internal.KeyFields))
			},
		},
		{
			name:          "none",
			existingCtxFn: nil,
			fieldSelector: "",
			checkContext: func(t *testing.T, ctx context.Context) {
				assert.Nil(t, ctx.Value(internal.KeyPodCounting))
				assert.Nil(t, ctx.Value(internal.KeyFields))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Call the chained function directly
			chainedFn := chainContext(tt.existingCtxFn, tt.fieldSelector)
			resultCtx := chainedFn(context.Background())
			tt.checkContext(t, resultCtx)
		})
	}
}
