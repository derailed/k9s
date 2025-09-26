// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"os"
	"testing"
)

// TestValidateTLSConfig_SecurityTests tests the validateTLSConfig function for security vulnerabilities
// SECURITY FIX (SEC-004): These tests ensure that TLS configuration is properly validated
func TestValidateTLSConfig_SecurityTests(t *testing.T) {
	tests := []struct {
		name        string
		insecure    bool
		envOverride string
		expectError bool
		description string
	}{
		{
			name:        "secure_tls_default",
			insecure:    false,
			envOverride: "",
			expectError: false,
			description: "Secure TLS configuration should be allowed by default",
		},
		{
			name:        "insecure_tls_without_override",
			insecure:    true,
			envOverride: "",
			expectError: true,
			description: "Insecure TLS should require environment variable override",
		},
		{
			name:        "insecure_tls_with_override",
			insecure:    true,
			envOverride: "true",
			expectError: false,
			description: "Insecure TLS should be allowed with environment variable override",
		},
		{
			name:        "insecure_tls_with_false_override",
			insecure:    true,
			envOverride: "false",
			expectError: true,
			description: "Insecure TLS should still be blocked with false override",
		},
		{
			name:        "insecure_tls_with_invalid_override",
			insecure:    true,
			envOverride: "invalid",
			expectError: true,
			description: "Insecure TLS should be blocked with invalid override value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envOverride != "" {
				os.Setenv("K9S_ALLOW_INSECURE_TLS", tt.envOverride)
				defer os.Unsetenv("K9S_ALLOW_INSECURE_TLS")
			}

			err := validateTLSConfig(tt.insecure)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for %s, but got none", tt.description)
				} else {
					t.Logf("✓ Correctly rejected insecure TLS: %v", err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error for %s, but got: %v", tt.description, err)
				} else {
					t.Logf("✓ Correctly allowed TLS configuration: %s", tt.description)
				}
			}
		})
	}
}
