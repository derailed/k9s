// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package ai

import (
	"bytes"
	"text/template"
)

// K8sContext holds Kubernetes context information for AI prompts.
type K8sContext struct {
	ClusterName      string
	ContextName      string
	Namespace        string
	ResourceType     string
	SelectedResource string
	ResourceYAML     string
	Events           string
}

const systemPromptTemplate = `You are a Kubernetes assistant integrated into k9s.
You have access to the following context:

Cluster: {{.ClusterName}}
Context: {{.ContextName}}
Namespace: {{.Namespace}}
Current View: {{.ResourceType}}
{{- if .SelectedResource}}
Selected Resource: {{.SelectedResource}}
{{- if .ResourceYAML}}
Resource YAML:
{{.ResourceYAML}}
{{- end}}
{{- end}}
{{- if .Events}}
Recent Events:
{{.Events}}
{{- end}}

Help the user understand and troubleshoot their Kubernetes resources.
Be concise and actionable. Suggest k9s commands when relevant (e.g., ":pods", ":logs", ":describe").
When providing solutions, explain the root cause first, then the fix.`

var systemTmpl = template.Must(template.New("system").Parse(systemPromptTemplate))

// BuildSystemPrompt builds the system prompt from the K8s context.
func BuildSystemPrompt(ctx *K8sContext) (string, error) {
	if ctx == nil {
		ctx = &K8sContext{}
	}

	var buf bytes.Buffer
	if err := systemTmpl.Execute(&buf, ctx); err != nil {
		return "", err
	}

	return buf.String(), nil
}
