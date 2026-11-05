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

// TestChainContextFunctions verifies that chainContextFunctions preserves all context keys
// in the correct order: existing context first, then field selector.
// This regression test ensures that issue #4139 (custom jump overwriting context) doesn't resurface.
//
// The 4 scenarios comprehensively cover all combinations:
// 1. Both existing context function and field selector present (both must survive)
// 2. Existing context function only (no field selector)
// 3. Field selector only (no existing context function)
// 4. Neither (clean chaining, no-op)
func TestChainContextFunctions(t *testing.T) {
	tests := []struct {
		name                string
		existingCtxFn       view.ContextFunc
		fieldSelector       string
		expectedContextKeys map[interface{}]interface{} // context key -> expected value
		description         string
	}{
		{
			name: "both_existing_and_field",
			description: "Chains existing context function + field selector. " +
				"Verifies that pod counting context (KeyPodCounting) is preserved alongside field filtering (KeyFields). " +
				"This is the critical case from issue #4139.",
			existingCtxFn: func(ctx context.Context) context.Context {
				// Simulate what Node or another viewer does: set KeyPodCounting
				ctx = context.WithValue(ctx, internal.KeyPodCounting, true)
				// Also set KeyFactory to test multiple context values
				ctx = context.WithValue(ctx, internal.KeyFactory, "node")
				return ctx
			},
			fieldSelector: "metadata.name=test-pod",
			expectedContextKeys: map[interface{}]interface{}{
				internal.KeyPodCounting: true,
				internal.KeyFactory:     "node",
				internal.KeyFields:      "metadata.name=test-pod",
			},
		},
		{
			name: "existing_only",
			description: "Only existing context function, no field selector. " +
				"Ensures that existing context is still applied even when there's no field selector to add. " +
				"Regression detector: if existingCtxFn call is removed, this test fails.",
			existingCtxFn: func(ctx context.Context) context.Context {
				ctx = context.WithValue(ctx, internal.KeyPodCounting, false)
				ctx = context.WithValue(ctx, internal.KeyFactory, "pod")
				return ctx
			},
			fieldSelector: "",
			expectedContextKeys: map[interface{}]interface{}{
				internal.KeyPodCounting: false,
				internal.KeyFactory:     "pod",
			},
		},
		{
			name: "field_only",
			description: "Only field selector, no existing context function. " +
				"Ensures nil-safety when existingCtxFn is nil. " +
				"Regression detector: if nil check is removed, this test would panic.",
			existingCtxFn:       nil,
			fieldSelector:       "metadata.namespace=default",
			expectedContextKeys: map[interface{}]interface{}{
				internal.KeyFields: "metadata.namespace=default",
			},
		},
		{
			name: "neither",
			description: "No existing context function, no field selector. " +
				"Edge case: ensures the chaining function handles empty inputs cleanly. " +
				"The returned function should be a valid no-op.",
			existingCtxFn:       nil,
			fieldSelector:       "",
			expectedContextKeys: map[interface{}]interface{}{}, // No keys expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a browser to get access to chainContextFunctions.
			// We're testing the core chaining logic, not the browser itself.
			b := view.NewBrowser(client.NewGVR("v1/pods"))
			require.NotNil(t, b)

			// Test the chaining function directly by calling it and then invoking the result.
			// This is what customJump does at lines 105-106.
			chainedFn := view.ChainContextFunctions(tt.existingCtxFn, tt.fieldSelector)
			require.NotNil(t, chainedFn, "chainContextFunctions should always return a function")

			// Execute the chained context function with a clean background context
			resultCtx := chainedFn(context.Background())
			require.NotNil(t, resultCtx)

			// Verify all expected context keys are present with correct values
			for key, expectedValue := range tt.expectedContextKeys {
				actualValue := resultCtx.Value(key)
				assert.Equal(t, expectedValue, actualValue,
					"context key %v: got %v, want %v", key, actualValue, expectedValue)
			}

			// Verify no unexpected keys are present
			// (for "neither" case, ensure we didn't accidentally add anything)
			if len(tt.expectedContextKeys) == 0 {
				// In the "neither" case, the context should be clean (no extra keys)
				// We can't exhaustively check all possible keys, but we can verify the main ones
				assert.Nil(t, resultCtx.Value(internal.KeyPodCounting))
				assert.Nil(t, resultCtx.Value(internal.KeyFields))
				assert.Nil(t, resultCtx.Value(internal.KeyFactory))
			}

			t.Logf("✓ %s: %s", tt.name, tt.description)
		})
	}
}

// TestChainContextFunctionsOrder verifies that context chaining respects the order:
// 1. Existing context function applies first
// 2. Field selector applies second (can't be overwritten by existing function)
// This test specifically detects if the order is reversed.
func TestChainContextFunctionsOrder(t *testing.T) {
	// Create an existing context function that sets KeyFields to a different value
	existingCtxFn := func(ctx context.Context) context.Context {
		// This simulates a scenario where existing context also sets KeyFields
		return context.WithValue(ctx, internal.KeyFields, "old-selector")
	}

	newFieldSelector := "new-selector"

	// Chain them
	chainedFn := view.ChainContextFunctions(existingCtxFn, newFieldSelector)
	resultCtx := chainedFn(context.Background())

	// The new field selector should overwrite the old one
	// (second assignment wins in context.WithValue usage)
	fieldsValue := resultCtx.Value(internal.KeyFields).(string)
	assert.Equal(t, newFieldSelector, fieldsValue,
		"new field selector should take precedence over existing context fields")
}

// BenchmarkChainContextFunctions measures the performance of context chaining.
// This ensures that the chaining helper doesn't introduce significant overhead.
func BenchmarkChainContextFunctions(b *testing.B) {
	existingFn := func(ctx context.Context) context.Context {
		return context.WithValue(ctx, internal.KeyPodCounting, true)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		chainedFn := view.ChainContextFunctions(existingFn, "test-selector")
		_ = chainedFn(context.Background())
	}
}
