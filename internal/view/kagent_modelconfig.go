// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// KModelConfig represents a kagent ModelConfig viewer.
type KModelConfig struct {
	ResourceViewer
}

// NewKModelConfig returns a new ModelConfig viewer.
func NewKModelConfig(gvr *client.GVR) ResourceViewer {
	m := KModelConfig{
		ResourceViewer: NewBrowser(gvr),
	}
	m.AddBindKeysFn(m.bindKeys)
	m.GetTable().SetEnterFn(m.showModelConfigDetails)

	return &m
}

func (m *KModelConfig) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyA:        ui.NewKeyAction("Show Agents", m.showAgentsCmd, true),
		ui.KeyS:        ui.NewKeyAction("Show Secret", m.showSecretCmd, true),
		tcell.KeyCtrlT: ui.NewKeyAction("Test Model", m.testModelCmd, true),
	})
}

func (m *KModelConfig) showModelConfigDetails(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	// Show YAML view by default
	v := NewLiveView(app, yamlAction, model.NewYAML(gvr, path))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

// showAgentsCmd shows all agents using this ModelConfig
func (m *KModelConfig) showAgentsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := m.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)
	m.App().Flash().Infof("Finding agents using ModelConfig %s...", name)

	// Navigate to agents view - users can filter from there
	agentPath := fmt.Sprintf("agents -n %s", ns)
	m.App().gotoResource(agentPath, "", false, true)

	return nil
}

// showSecretCmd shows the associated API key secret
func (m *KModelConfig) showSecretCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := m.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the ModelConfig object
	ctx := context.Background()
	o, err := m.App().factory.Get(m.GVR(), path, true, nil)
	if err != nil {
		m.App().Flash().Errf("Failed to get ModelConfig: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	// Extract secret reference from spec
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	secretName, ok := spec["apiKeySecret"].(string)
	if !ok || secretName == "" {
		m.App().Flash().Warn("ModelConfig has no API key secret configured")
		return nil
	}

	// Navigate to the secret
	secretPath := fmt.Sprintf("secrets %s/%s", raw.GetNamespace(), secretName)
	m.App().gotoResource(secretPath, "", false, true)

	_ = ctx
	return nil
}

// testModelCmd provides info about testing the model
func (m *KModelConfig) testModelCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := m.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the ModelConfig object
	o, err := m.App().factory.Get(m.GVR(), path, true, nil)
	if err != nil {
		m.App().Flash().Errf("Failed to get ModelConfig: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	// Extract model info
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	provider := ""
	if p, ok := spec["provider"].(string); ok {
		provider = p
	}
	model := ""
	if m, ok := spec["model"].(string); ok {
		model = m
	}

	// Build info message
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("ModelConfig: %s\n\n", raw.GetName()))
	sb.WriteString(fmt.Sprintf("Provider: %s\n", provider))
	sb.WriteString(fmt.Sprintf("Model: %s\n\n", model))

	// Provider-specific info
	switch provider {
	case "OpenAI":
		if openai, ok := spec["openAI"].(map[string]interface{}); ok {
			if baseURL, ok := openai["baseUrl"].(string); ok && baseURL != "" {
				sb.WriteString(fmt.Sprintf("Base URL: %s\n", baseURL))
			}
			if temp, ok := openai["temperature"].(string); ok && temp != "" {
				sb.WriteString(fmt.Sprintf("Temperature: %s\n", temp))
			}
		}
	case "Anthropic":
		if anthropic, ok := spec["anthropic"].(map[string]interface{}); ok {
			if baseURL, ok := anthropic["baseUrl"].(string); ok && baseURL != "" {
				sb.WriteString(fmt.Sprintf("Base URL: %s\n", baseURL))
			}
			if maxTokens, ok := anthropic["maxTokens"].(int64); ok {
				sb.WriteString(fmt.Sprintf("Max Tokens: %d\n", maxTokens))
			}
		}
	case "AzureOpenAI":
		if azure, ok := spec["azureOpenAI"].(map[string]interface{}); ok {
			if endpoint, ok := azure["azureEndpoint"].(string); ok && endpoint != "" {
				sb.WriteString(fmt.Sprintf("Endpoint: %s\n", endpoint))
			}
			if deployment, ok := azure["azureDeployment"].(string); ok && deployment != "" {
				sb.WriteString(fmt.Sprintf("Deployment: %s\n", deployment))
			}
		}
	case "Ollama":
		if ollama, ok := spec["ollama"].(map[string]interface{}); ok {
			if host, ok := ollama["host"].(string); ok && host != "" {
				sb.WriteString(fmt.Sprintf("Host: %s\n", host))
			}
		}
	}

	sb.WriteString("\nTo test this model, create an agent using this ModelConfig")
	sb.WriteString("\nand invoke it with 'kagent invoke <agent-name>'")

	// Show in details view
	v := NewDetails(m.App(), "ModelConfig Info", raw.GetName(), contentTXT, true)
	v.Update(sb.String())
	if err := m.App().inject(v, false); err != nil {
		m.App().Flash().Err(err)
	}

	return nil
}

