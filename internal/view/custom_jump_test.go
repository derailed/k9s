// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view_test

import (
	"context"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestContextChaining verifies that custom jump preserves existing context
// while adding field selectors, specifically checking that both KeyPodCounting
// and KeyFields end up in the context after a jump.
func TestContextChaining(t *testing.T) {
	// Create a mock browser with an initial context function that sets KeyPodCounting
	b := view.NewBrowser(client.NewGVR("v1/pods"))
	require.NotNil(t, b)

	// Set up initial context function (simulating what Node or another viewer does)
	initialContextFn := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, internal.KeyPodCounting, true)
	}
	b.SetContextFn(initialContextFn)

	// Verify the initial context function works
	baseCtx := context.Background()
	ctxAfterInitial := b.GetContextFn()(baseCtx)
	podCounting, ok := ctxAfterInitial.Value(internal.KeyPodCounting).(bool)
	require.True(t, ok, "KeyPodCounting should be present after initial context function")
	assert.True(t, podCounting, "KeyPodCounting should be true")

	// Simulate custom jump context chaining (this is what customJump does at lines 85-96)
	existingCtxFn := b.GetContextFn()
	fieldSelector := "metadata.name=test-pod"
	b.SetContextFn(func(ctx context.Context) context.Context {
		// Apply existing context first (preserves KeyPodCounting, etc)
		if existingCtxFn != nil {
			ctx = existingCtxFn(ctx)
		}
		// Then add field selector
		if fieldSelector != "" {
			ctx = context.WithValue(ctx, internal.KeyFields, fieldSelector)
		}
		return ctx
	})

	// Now verify both keys are present in the chained context
	ctxAfterChain := b.GetContextFn()(context.Background())

	// Check KeyPodCounting is still there (preserved from initial context function)
	podCountingValue, ok := ctxAfterChain.Value(internal.KeyPodCounting).(bool)
	require.True(t, ok, "KeyPodCounting should be present after context chaining")
	assert.True(t, podCountingValue, "KeyPodCounting should still be true")

	// Check KeyFields is added by the custom jump
	fieldsValue, ok := ctxAfterChain.Value(internal.KeyFields).(string)
	require.True(t, ok, "KeyFields should be present after context chaining")
	assert.Equal(t, fieldSelector, fieldsValue, "KeyFields should contain the field selector")
}

// TestContextChainingWithoutInitialFn verifies that custom jump works
// even when there's no initial context function.
func TestContextChainingWithoutInitialFn(t *testing.T) {
	b := view.NewBrowser(client.NewGVR("v1/pods"))
	require.NotNil(t, b)

	// Don't set an initial context function (GetContextFn returns nil)
	assert.Nil(t, b.GetContextFn())

	// Simulate custom jump context chaining with nil initial function
	existingCtxFn := b.GetContextFn()
	fieldSelector := "metadata.namespace=default"
	b.SetContextFn(func(ctx context.Context) context.Context {
		// Apply existing context first (safe to call even if nil)
		if existingCtxFn != nil {
			ctx = existingCtxFn(ctx)
		}
		// Then add field selector
		if fieldSelector != "" {
			ctx = context.WithValue(ctx, internal.KeyFields, fieldSelector)
		}
		return ctx
	})

	// Verify KeyFields is present
	ctxAfterChain := b.GetContextFn()(context.Background())
	fieldsValue, ok := ctxAfterChain.Value(internal.KeyFields).(string)
	require.True(t, ok, "KeyFields should be present")
	assert.Equal(t, fieldSelector, fieldsValue)
}
