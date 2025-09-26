// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"encoding/base64"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

// TestExtractSecrets_SecurityTests tests the ExtractSecrets function for security vulnerabilities
// SECURITY FIX (SEC-003): These tests ensure that secrets are not accidentally exposed in plaintext
func TestExtractSecrets_SecurityTests(t *testing.T) {
	// Create a test secret with sensitive data
	testSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte(base64.StdEncoding.EncodeToString([]byte("admin"))),
			"password": []byte(base64.StdEncoding.EncodeToString([]byte("secretpassword123"))),
			"api-key":  []byte(base64.StdEncoding.EncodeToString([]byte("sk-1234567890abcdef"))),
		},
	}

	// Convert to unstructured for testing
	unstructuredSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testSecret)
	if err != nil {
		t.Fatalf("Failed to convert secret to unstructured: %v", err)
	}

	tests := []struct {
		name        string
		secret      runtime.Object
		expectError bool
		description string
	}{
		{
			name:        "valid_secret_returns_encoded",
			secret:      &unstructured.Unstructured{Object: unstructuredSecret},
			expectError: false,
			description: "Valid secret should return encoded data by default",
		},
		{
			name:        "invalid_object_type",
			secret:      &v1.Pod{}, // Wrong type
			expectError: true,
			description: "Invalid object type should return error",
		},
		{
			name:        "nil_secret",
			secret:      nil,
			expectError: true,
			description: "Nil secret should return error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractSecrets(tt.secret)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for secret %v, but got none", tt.secret)
				}
				t.Logf("✓ Correctly rejected invalid secret: %v -> %v", tt.secret, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for secret %v: %v", tt.secret, err)
				}

				// Verify that the returned data is Base64-encoded (not decoded)
				if result != nil {
					for key, value := range result {
						// The value should be Base64-encoded, not plaintext
						_, err := base64.StdEncoding.DecodeString(value)
						if err != nil {
							t.Errorf("Value for key %s is not Base64-encoded: %v", key, err)
						}

						// Verify it's not the original plaintext
						originalValue := testSecret.Data[key]
						if string(originalValue) != value {
							t.Errorf("Expected encoded value %s, got %s", string(originalValue), value)
						}

						t.Logf("✓ Correctly returned encoded data for key %s: %s", key, value)
					}
				}
			}
		})
	}
}

// TestExtractSecretsDecoded_SecurityTests tests the ExtractSecretsDecoded function
func TestExtractSecretsDecoded_SecurityTests(t *testing.T) {
	// Create a test secret with sensitive data
	testSecret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"username": []byte(base64.StdEncoding.EncodeToString([]byte("admin"))),
			"password": []byte(base64.StdEncoding.EncodeToString([]byte("secretpassword123"))),
			"api-key":  []byte(base64.StdEncoding.EncodeToString([]byte("sk-1234567890abcdef"))),
		},
	}

	// Convert to unstructured for testing
	unstructuredSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(testSecret)
	if err != nil {
		t.Fatalf("Failed to convert secret to unstructured: %v", err)
	}

	tests := []struct {
		name          string
		secret        runtime.Object
		userConfirmed bool
		expectError   bool
		expectDecoded bool
		description   string
	}{
		{
			name:          "decoded_with_confirmation",
			secret:        &unstructured.Unstructured{Object: unstructuredSecret},
			userConfirmed: true,
			expectError:   false,
			expectDecoded: true,
			description:   "Secret should be decoded when user confirms",
		},
		{
			name:          "encoded_without_confirmation",
			secret:        &unstructured.Unstructured{Object: unstructuredSecret},
			userConfirmed: false,
			expectError:   true,
			expectDecoded: false,
			description:   "Secret should not be decoded without user confirmation",
		},
		{
			name:          "invalid_object_type",
			secret:        &v1.Pod{}, // Wrong type
			userConfirmed: true,
			expectError:   true,
			expectDecoded: false,
			description:   "Invalid object type should return error even with confirmation",
		},
		{
			name:          "nil_secret",
			secret:        nil,
			userConfirmed: true,
			expectError:   true,
			expectDecoded: false,
			description:   "Nil secret should return error even with confirmation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractSecretsDecoded(tt.secret, tt.userConfirmed)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for secret %v with confirmation %v, but got none", tt.secret, tt.userConfirmed)
				}
				t.Logf("✓ Correctly rejected secret access: %v -> %v", tt.secret, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for secret %v with confirmation %v: %v", tt.secret, tt.userConfirmed, err)
				}

				if tt.expectDecoded && result != nil {
					// Verify that the returned data is decoded (plaintext)
					for key, value := range result {
						// The value should be decoded plaintext, not Base64-encoded
						_, err := base64.StdEncoding.DecodeString(value)
						if err == nil {
							t.Errorf("Value for key %s should be decoded plaintext, but appears to be Base64-encoded: %s", key, value)
						}

						// Verify it matches the expected decoded value
						expectedValue := string(testSecret.Data[key])
						decodedExpected, _ := base64.StdEncoding.DecodeString(expectedValue)
						if string(decodedExpected) != value {
							t.Errorf("Expected decoded value %s, got %s", string(decodedExpected), value)
						}

						t.Logf("✓ Correctly returned decoded data for key %s: %s", key, value)
					}
				}
			}
		})
	}
}

// TestExtractSecrets_EdgeCases tests edge cases and boundary conditions
func TestExtractSecrets_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		secret      *v1.Secret
		expectError bool
		description string
	}{
		{
			name: "empty_secret",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "empty-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{},
			},
			expectError: false,
			description: "Empty secret should be handled gracefully",
		},
		{
			name: "secret_with_invalid_base64",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "invalid-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"invalid": []byte("not-base64-encoded"),
				},
			},
			expectError: false,
			description: "Secret with invalid Base64 should be handled gracefully",
		},
		{
			name: "secret_with_large_data",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "large-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"large-key": []byte(base64.StdEncoding.EncodeToString(make([]byte, 10000))),
				},
			},
			expectError: false,
			description: "Secret with large data should be handled gracefully",
		},
		{
			name: "secret_with_special_characters",
			secret: &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "special-secret",
					Namespace: "default",
				},
				Data: map[string][]byte{
					"special-chars": []byte(base64.StdEncoding.EncodeToString([]byte("!@#$%^&*()_+-=[]{}|;':\",./<>?"))),
					"unicode":       []byte(base64.StdEncoding.EncodeToString([]byte("测试中文"))),
					"newlines":      []byte(base64.StdEncoding.EncodeToString([]byte("line1\nline2\nline3"))),
				},
			},
			expectError: false,
			description: "Secret with special characters should be handled gracefully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to unstructured for testing
			unstructuredSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(tt.secret)
			if err != nil {
				t.Fatalf("Failed to convert secret to unstructured: %v", err)
			}

			result, err := ExtractSecrets(&unstructured.Unstructured{Object: unstructuredSecret})

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for secret %v, but got none", tt.secret.Name)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for secret %v: %v", tt.secret.Name, err)
				}

				// Verify the result structure
				if result == nil {
					t.Errorf("Expected non-nil result for secret %v", tt.secret.Name)
				}

				// Verify all keys are present
				for key := range tt.secret.Data {
					if _, exists := result[key]; !exists {
						t.Errorf("Expected key %s to be present in result", key)
					}
				}
			}
		})
	}
}

// TestExtractSecrets_RealWorldScenarios tests real-world secret scenarios
func TestExtractSecrets_RealWorldScenarios(t *testing.T) {
	scenarios := []struct {
		name        string
		secretType  string
		secretData  map[string][]byte
		description string
	}{
		{
			name:       "database_credentials",
			secretType: "Opaque",
			secretData: map[string][]byte{
				"username": []byte(base64.StdEncoding.EncodeToString([]byte("dbuser"))),
				"password": []byte(base64.StdEncoding.EncodeToString([]byte("dbpass123"))),
				"host":     []byte(base64.StdEncoding.EncodeToString([]byte("db.example.com"))),
				"port":     []byte(base64.StdEncoding.EncodeToString([]byte("5432"))),
			},
			description: "Database credentials secret should be handled securely",
		},
		{
			name:       "api_keys",
			secretType: "Opaque",
			secretData: map[string][]byte{
				"api-key":    []byte(base64.StdEncoding.EncodeToString([]byte("sk-1234567890abcdef"))),
				"secret-key": []byte(base64.StdEncoding.EncodeToString([]byte("sk_secret_abcdef123456"))),
				"webhook":    []byte(base64.StdEncoding.EncodeToString([]byte("whsec_1234567890abcdef"))),
			},
			description: "API keys secret should be handled securely",
		},
		{
			name:       "tls_certificates",
			secretType: "kubernetes.io/tls",
			secretData: map[string][]byte{
				"tls.crt": []byte(base64.StdEncoding.EncodeToString([]byte("-----BEGIN CERTIFICATE-----\nMII...\n-----END CERTIFICATE-----"))),
				"tls.key": []byte(base64.StdEncoding.EncodeToString([]byte("-----BEGIN PRIVATE KEY-----\nMII...\n-----END PRIVATE KEY-----"))),
			},
			description: "TLS certificates secret should be handled securely",
		},
		{
			name:       "docker_registry",
			secretType: "kubernetes.io/dockerconfigjson",
			secretData: map[string][]byte{
				".dockerconfigjson": []byte(base64.StdEncoding.EncodeToString([]byte(`{"auths":{"registry.example.com":{"username":"user","password":"pass","auth":"dXNlcjpwYXNz"}}}`))),
			},
			description: "Docker registry secret should be handled securely",
		},
	}

	for _, scenario := range scenarios {
		t.Run(scenario.name, func(t *testing.T) {
			secret := &v1.Secret{
				ObjectMeta: metav1.ObjectMeta{
					Name:      scenario.name,
					Namespace: "default",
				},
				Type: v1.SecretType(scenario.secretType),
				Data: scenario.secretData,
			}

			// Convert to unstructured for testing
			unstructuredSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
			if err != nil {
				t.Fatalf("Failed to convert secret to unstructured: %v", err)
			}

			// Test encoded extraction (default behavior)
			encodedResult, err := ExtractSecrets(&unstructured.Unstructured{Object: unstructuredSecret})
			if err != nil {
				t.Errorf("Failed to extract encoded secrets: %v", err)
			}

			// Verify all keys are present and encoded
			for key, expectedValue := range scenario.secretData {
				if actualValue, exists := encodedResult[key]; !exists {
					t.Errorf("Expected key %s to be present in encoded result", key)
				} else if actualValue != string(expectedValue) {
					t.Errorf("Expected encoded value %s, got %s", string(expectedValue), actualValue)
				}
			}

			// Test decoded extraction with confirmation
			decodedResult, err := ExtractSecretsDecoded(&unstructured.Unstructured{Object: unstructuredSecret}, true)
			if err != nil {
				t.Errorf("Failed to extract decoded secrets: %v", err)
			}

			// Verify all keys are present and decoded
			for key, expectedEncodedValue := range scenario.secretData {
				if actualValue, exists := decodedResult[key]; !exists {
					t.Errorf("Expected key %s to be present in decoded result", key)
				} else {
					// Decode the expected value to compare
					expectedDecoded, err := base64.StdEncoding.DecodeString(string(expectedEncodedValue))
					if err != nil {
						t.Errorf("Failed to decode expected value for key %s: %v", key, err)
					} else if actualValue != string(expectedDecoded) {
						t.Errorf("Expected decoded value %s, got %s", string(expectedDecoded), actualValue)
					}
				}
			}

			t.Logf("✓ Correctly handled %s secret scenario", scenario.name)
		})
	}
}

// BenchmarkExtractSecrets benchmarks the secret extraction functions
func BenchmarkExtractSecrets(b *testing.B) {
	// Create a test secret
	secret := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "benchmark-secret",
			Namespace: "default",
		},
		Data: map[string][]byte{
			"key1": []byte(base64.StdEncoding.EncodeToString([]byte("value1"))),
			"key2": []byte(base64.StdEncoding.EncodeToString([]byte("value2"))),
			"key3": []byte(base64.StdEncoding.EncodeToString([]byte("value3"))),
		},
	}

	unstructuredSecret, _ := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	obj := &unstructured.Unstructured{Object: unstructuredSecret}

	b.Run("ExtractSecrets", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ExtractSecrets(obj)
		}
	})

	b.Run("ExtractSecretsDecoded", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = ExtractSecretsDecoded(obj, true)
		}
	})
}
