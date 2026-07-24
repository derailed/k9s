// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package cmd

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

// fatalConfigError flags a non-recoverable kubeconfig loading failure.
// When the kubeconfig cannot be parsed/loaded at all, k9s cannot obtain any
// context and entering the TUI is pointless (it would only loop on the same
// error). Such errors must be surfaced to the user on the terminal.
type fatalConfigError struct {
	path string
	err  error
}

func (e *fatalConfigError) Error() string {
	if e.path != "" {
		return fmt.Sprintf("kubeconfig %q could not be loaded: %v", e.path, e.err)
	}

	return fmt.Sprintf("kubeconfig could not be loaded: %v", e.err)
}

func (e *fatalConfigError) Unwrap() error { return e.err }

// UserMessage returns a human friendly, multi-line description of the failure
// suitable for direct terminal output.
func (e *fatalConfigError) UserMessage() string {
	var b strings.Builder
	b.WriteString("Unable to load your Kubernetes configuration (kubeconfig).\n")
	if e.path != "" {
		b.WriteString(fmt.Sprintf("File: %s\n", e.path))
	}
	b.WriteString(fmt.Sprintf("Reason: %s\n", rootCause(e.err)))
	b.WriteString("\nK9s cannot start until this is resolved. Please fix your kubeconfig and try again.")

	return b.String()
}

// checkFatalConfigError probes whether the kubeconfig is fundamentally
// unusable. It returns a *fatalConfigError only when the kubeconfig cannot be
// loaded/parsed at all (e.g. malformed file, duplicate entries). It explicitly
// does NOT treat a parseable-but-unreachable cluster as fatal so users keep the
// ability to switch contexts from within the TUI.
func checkFatalConfigError(cfg *client.Config) *fatalConfigError {
	if cfg == nil {
		return nil
	}
	// RawConfig is the lowest-level kubeconfig load. If it fails, no context can
	// be resolved and the file is broken at the parse/merge level.
	if _, err := cfg.RawConfig(); err != nil {
		return &fatalConfigError{path: kubeconfigPath(cfg), err: err}
	}

	return nil
}

// kubeconfigPath best-effort resolves the kubeconfig file path for messaging.
func kubeconfigPath(cfg *client.Config) string {
	acc, err := cfg.ConfigAccess()
	if err != nil || acc == nil {
		return ""
	}
	if f := acc.GetExplicitFile(); f != "" {
		return f
	}
	if pp := acc.GetLoadingPrecedence(); len(pp) > 0 {
		return strings.Join(pp, ", ")
	}

	return ""
}

// rootCause walks the error chain and returns the deepest message, stripping
// k9s' own wrapping so the user sees the underlying client-go cause.
func rootCause(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	// errors.Join concatenates messages with newlines; surface the first line
	// that mentions the config load failure when present, else the whole thing.
	for _, line := range strings.Split(msg, "\n") {
		if strings.Contains(line, "error loading config file") ||
			strings.Contains(line, "duplicate") {
			return strings.TrimSpace(line)
		}
	}

	return strings.TrimSpace(msg)
}
