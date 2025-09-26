// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"os"
	"testing"
)

// TestValidatePrivilegedPodCreation_SecurityTests tests the validatePrivilegedPodCreation function for security vulnerabilities
// SECURITY FIX (SEC-005): These tests ensure that privileged pod creation requires proper validation
func TestValidatePrivilegedPodCreation_SecurityTests(t *testing.T) {
	tests := []struct {
		name        string
		nodeName    string
		envOverride string
		expectError bool
		description string
	}{
		{
			name:        "privileged_pod_without_override",
			nodeName:    "test-node",
			envOverride: "",
			expectError: true,
			description: "Privileged pod should require environment variable override",
		},
		{
			name:        "privileged_pod_with_override",
			nodeName:    "test-node",
			envOverride: "true",
			expectError: false,
			description: "Privileged pod should be allowed with environment variable override",
		},
		{
			name:        "privileged_pod_with_false_override",
			nodeName:    "test-node",
			envOverride: "false",
			expectError: true,
			description: "Privileged pod should still be blocked with false override",
		},
		{
			name:        "privileged_pod_with_invalid_override",
			nodeName:    "test-node",
			envOverride: "invalid",
			expectError: true,
			description: "Privileged pod should be blocked with invalid override value",
		},
		{
			name:        "empty_node_name",
			nodeName:    "",
			envOverride: "true",
			expectError: false,
			description: "Empty node name should be allowed with override",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envOverride != "" {
				os.Setenv("K9S_ALLOW_PRIVILEGED_PODS", tt.envOverride)
				defer os.Unsetenv("K9S_ALLOW_PRIVILEGED_PODS")
			}

			err := validatePrivilegedPodCreation(tt.nodeName)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.description)
				} else {
					t.Logf("✓ Correctly rejected privileged pod: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, but got: %v", tt.description, err)
				} else {
					t.Logf("✓ Correctly allowed pod creation: %s", tt.description)
				}
			}
		})
	}
}
