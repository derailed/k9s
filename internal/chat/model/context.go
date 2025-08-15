// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
)

// K9sContext represents the current k9s context for LLM awareness.
type K9sContext struct {
	ClusterName     string
	ContextName     string
	Namespace       string
	CurrentResource string
	CurrentView     string
	SelectedItem    string
}

// NewK9sContext creates a new k9s context.
func NewK9sContext() *K9sContext {
	return &K9sContext{}
}

// Update updates the context with current k9s state.
func (ctx *K9sContext) Update(clusterName, contextName, namespace, resource, view, selectedItem string) {
	ctx.ClusterName = clusterName
	ctx.ContextName = contextName
	ctx.Namespace = namespace
	ctx.CurrentResource = resource
	ctx.CurrentView = view
	ctx.SelectedItem = selectedItem
}

// GetSystemPrompt generates a system prompt with current k9s context.
func (ctx *K9sContext) GetSystemPrompt() string {
	var parts []string

	parts = append(parts, "You are an expert Kubernetes assistant integrated into k9s, a Kubernetes CLI tool.")
	parts = append(parts, "You help users understand and manage their Kubernetes clusters safely and effectively.")
	parts = append(parts, "")

	// Current context information
	parts = append(parts, "CURRENT CONTEXT:")
	if ctx.ClusterName != "" && ctx.ClusterName != client.NA {
		parts = append(parts, fmt.Sprintf("- Cluster: %s", ctx.ClusterName))
	}
	if ctx.ContextName != "" && ctx.ContextName != client.NA {
		parts = append(parts, fmt.Sprintf("- Context: %s", ctx.ContextName))
	}
	if ctx.Namespace != "" && ctx.Namespace != client.BlankNamespace {
		parts = append(parts, fmt.Sprintf("- Namespace: %s", ctx.Namespace))
	} else {
		parts = append(parts, "- Namespace: all namespaces")
	}
	if ctx.CurrentResource != "" {
		parts = append(parts, fmt.Sprintf("- Current Resource: %s", ctx.CurrentResource))
	}
	if ctx.CurrentView != "" {
		parts = append(parts, fmt.Sprintf("- Current View: %s", ctx.CurrentView))
	}
	if ctx.SelectedItem != "" {
		parts = append(parts, fmt.Sprintf("- Selected Item: %s", ctx.SelectedItem))
	}

	parts = append(parts, "")
	parts = append(parts, "GUIDELINES:")
	parts = append(parts, "- Provide clear, actionable advice for Kubernetes operations")
	parts = append(parts, "- Include relevant kubectl commands when helpful")
	parts = append(parts, "- Consider the current context when providing suggestions")
	parts = append(parts, "- Use markdown formatting for better readability")
	parts = append(parts, "- For shell commands, wrap them in ```bash code blocks")
	parts = append(parts, "- Explain the purpose and impact of suggested commands")
	parts = append(parts, "- Always prioritize safety and best practices")
	parts = append(parts, "- Be concise but comprehensive in your responses")

	return strings.Join(parts, "\n")
}

// GetSystemContext returns the system context for LLM providers.
func (ctx *K9sContext) GetSystemContext() string {
	return ctx.GetSystemPrompt()
}

// GetContextSummary returns a brief summary of the current context.
func (ctx *K9sContext) GetContextSummary() string {
	var parts []string

	if ctx.ContextName != "" && ctx.ContextName != client.NA {
		parts = append(parts, ctx.ContextName)
	}
	if ctx.Namespace != "" && ctx.Namespace != client.BlankNamespace {
		parts = append(parts, ctx.Namespace)
	}
	if ctx.CurrentResource != "" {
		parts = append(parts, ctx.CurrentResource)
	}

	if len(parts) == 0 {
		return "No context"
	}

	return strings.Join(parts, " | ")
}
