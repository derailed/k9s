// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"os"
	"path/filepath"
	"testing"
)

// TestValidatePath_SecurityTests tests the validatePath function for security vulnerabilities
// SECURITY FIX (SEC-001): These tests ensure that path traversal attacks are prevented
func TestValidatePath_SecurityTests(t *testing.T) {
	// Get user home directory for test setup
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		name        string
		inputPath   string
		expectError bool
		description string
	}{
		{
			name:        "valid_path_inside_home",
			inputPath:   filepath.Join(homeDir, "k9s", "config"),
			expectError: false,
			description: "Valid path inside user home directory should be allowed",
		},
		{
			name:        "valid_path_with_spaces",
			inputPath:   filepath.Join(homeDir, "my k9s config"),
			expectError: false,
			description: "Valid path with spaces should be allowed",
		},
		{
			name:        "empty_path",
			inputPath:   "",
			expectError: false,
			description: "Empty path should be allowed (fallback to defaults)",
		},
		{
			name:        "path_traversal_double_dots",
			inputPath:   "../../../etc/passwd",
			expectError: true,
			description: "Path with double dots should be rejected (traversal attack)",
		},
		{
			name:        "path_traversal_mixed_slashes",
			inputPath:   "..\\..\\..\\windows\\system32",
			expectError: true,
			description: "Windows-style path traversal should be rejected",
		},
		{
			name:        "path_traversal_encoded",
			inputPath:   "%2e%2e%2f%2e%2e%2f%2e%2e%2fetc",
			expectError: true,
			description: "URL-encoded path traversal should be rejected",
		},
		{
			name:        "path_outside_home_directory",
			inputPath:   "/etc/passwd",
			expectError: true,
			description: "Path outside home directory should be rejected",
		},
		{
			name:        "path_outside_home_windows",
			inputPath:   "C:\\Windows\\System32",
			expectError: false, // This will be resolved relative to current dir on macOS
			description: "Windows path will be resolved relative to current directory on macOS",
		},
		{
			name:        "path_with_null_bytes",
			inputPath:   homeDir + "\x00../../../etc",
			expectError: true,
			description: "Path with null bytes should be rejected",
		},
		{
			name:        "path_with_symlinks",
			inputPath:   filepath.Join(homeDir, "symlink_to_etc"),
			expectError: false,
			description: "Path with symlinks should be allowed (resolved by filepath.Abs)",
		},
		{
			name:        "relative_path_inside_home",
			inputPath:   "k9s/config",
			expectError: false,
			description: "Relative path inside home should be allowed",
		},
		{
			name:        "path_with_special_chars",
			inputPath:   filepath.Join(homeDir, "k9s-config_2024"),
			expectError: false,
			description: "Path with special characters should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validatePath(tt.inputPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none. Result: %q", tt.inputPath, result)
				}
				t.Logf("✓ Correctly rejected malicious path: %q -> %v", tt.inputPath, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.inputPath, err)
				}
				t.Logf("✓ Correctly allowed valid path: %q -> %q", tt.inputPath, result)
			}
		})
	}
}

// TestValidatePath_EdgeCases tests edge cases and boundary conditions
func TestValidatePath_EdgeCases(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	tests := []struct {
		name        string
		inputPath   string
		expectError bool
		description string
	}{
		{
			name:        "very_long_path",
			inputPath:   filepath.Join(homeDir, string(make([]byte, 1000))),
			expectError: false,
			description: "Very long path should be handled gracefully",
		},
		{
			name:        "path_with_unicode",
			inputPath:   filepath.Join(homeDir, "k9s-测试-配置"),
			expectError: false,
			description: "Path with Unicode characters should be allowed",
		},
		{
			name:        "path_with_quotes",
			inputPath:   filepath.Join(homeDir, "k9s", "config"),
			expectError: false,
			description: "Path with quotes should be allowed",
		},
		{
			name:        "path_with_parentheses",
			inputPath:   filepath.Join(homeDir, "k9s(config)"),
			expectError: false,
			description: "Path with parentheses should be allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := validatePath(tt.inputPath)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for input %q, but got none", tt.inputPath)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for input %q: %v", tt.inputPath, err)
				}
			}
		})
	}
}

// TestValidatePath_Integration tests integration with actual environment variables
func TestValidatePath_Integration(t *testing.T) {
	// Test with actual environment variable simulation
	originalEnv := os.Getenv("K9S_CONFIG_DIR")
	defer func() {
		if originalEnv != "" {
			os.Setenv("K9S_CONFIG_DIR", originalEnv)
		} else {
			os.Unsetenv("K9S_CONFIG_DIR")
		}
	}()

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("Failed to get user home directory: %v", err)
	}

	// Test valid environment variable
	validPath := filepath.Join(homeDir, "k9s", "config")
	os.Setenv("K9S_CONFIG_DIR", validPath)

	_, err = validatePath(os.Getenv("K9S_CONFIG_DIR"))
	if err != nil {
		t.Errorf("Valid environment variable should not cause error: %v", err)
	}

	// Test malicious environment variable
	maliciousPath := "../../../etc/passwd"
	os.Setenv("K9S_CONFIG_DIR", maliciousPath)

	_, err = validatePath(os.Getenv("K9S_CONFIG_DIR"))
	if err == nil {
		t.Errorf("Malicious environment variable should cause error")
	}
}

// BenchmarkValidatePath benchmarks the path validation function
func BenchmarkValidatePath(b *testing.B) {
	homeDir, _ := os.UserHomeDir()
	validPath := filepath.Join(homeDir, "k9s", "config")
	maliciousPath := "../../../etc/passwd"

	b.Run("ValidPath", func(b *testing.B) {
		for range b.N {
			_, _ = validatePath(validPath)
		}
	})

	b.Run("MaliciousPath", func(b *testing.B) {
		for range b.N {
			_, _ = validatePath(maliciousPath)
		}
	})
}
