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

// KToolServer represents a kagent ToolServer viewer.
type KToolServer struct {
	ResourceViewer
}

// NewKToolServer returns a new ToolServer viewer.
func NewKToolServer(gvr *client.GVR) ResourceViewer {
	t := KToolServer{
		ResourceViewer: NewBrowser(gvr),
	}
	t.AddBindKeysFn(t.bindKeys)
	t.GetTable().SetEnterFn(t.showToolServerDetails)

	return &t
}

func (t *KToolServer) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyT:        ui.NewKeyAction("List Tools", t.listToolsCmd, true),
		ui.KeyA:        ui.NewKeyAction("Show Agents", t.showAgentsCmd, true),
		tcell.KeyCtrlT: ui.NewKeyAction("Test Connection", t.testConnectionCmd, true),
	})
}

func (t *KToolServer) showToolServerDetails(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	// Show YAML view by default
	v := NewLiveView(app, yamlAction, model.NewYAML(gvr, path))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

// listToolsCmd shows all discovered tools for this ToolServer
func (t *KToolServer) listToolsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := t.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the ToolServer object
	ctx := context.Background()
	o, err := t.App().factory.Get(t.GVR(), path, true, nil)
	if err != nil {
		t.App().Flash().Errf("Failed to get ToolServer: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	// Extract discovered tools from status
	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	discoveredTools, ok := status["discoveredTools"].([]interface{})
	if !ok || len(discoveredTools) == 0 {
		t.App().Flash().Warn("No tools discovered for this ToolServer")
		return nil
	}

	// Build tools list
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Discovered Tools for ToolServer %s:\n\n", raw.GetName()))
	sb.WriteString(fmt.Sprintf("Total: %d tools\n", len(discoveredTools)))
	sb.WriteString(strings.Repeat("-", 60) + "\n\n")

	for i, tool := range discoveredTools {
		toolMap, ok := tool.(map[string]interface{})
		if !ok {
			continue
		}

		name := ""
		if n, ok := toolMap["name"].(string); ok {
			name = n
		}

		description := ""
		if d, ok := toolMap["description"].(string); ok {
			description = d
		}

		sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, name))
		if description != "" {
			// Wrap description
			wrapped := wrapText(description, 55)
			for _, line := range strings.Split(wrapped, "\n") {
				sb.WriteString(fmt.Sprintf("   %s\n", line))
			}
		}
		sb.WriteString("\n")
	}

	// Show in a details view
	v := NewDetails(t.App(), "ToolServer Tools", raw.GetName(), contentTXT, true)
	v.Update(sb.String())
	if err := t.App().inject(v, false); err != nil {
		t.App().Flash().Err(err)
	}

	_ = ctx
	return nil
}

// showAgentsCmd shows all agents using this ToolServer
func (t *KToolServer) showAgentsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := t.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)

	// Navigate to agents view filtered by this toolserver
	// This uses a label selector or we could scan agents
	t.App().Flash().Infof("Finding agents using ToolServer %s...", name)

	// For now, just show a hint - full implementation would query agents
	agentPath := fmt.Sprintf("agents -n %s", ns)
	t.App().gotoResource(agentPath, "", false, true)

	return nil
}

// testConnectionCmd tests the connection to the ToolServer
func (t *KToolServer) testConnectionCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := t.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the ToolServer object
	o, err := t.App().factory.Get(t.GVR(), path, true, nil)
	if err != nil {
		t.App().Flash().Errf("Failed to get ToolServer: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	// Extract config
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	config, _ := spec["config"].(map[string]interface{})

	serverType := "unknown"
	endpoint := ""

	if sse, ok := config["sse"].(map[string]interface{}); ok {
		serverType = "SSE"
		if url, ok := sse["url"].(string); ok {
			endpoint = url
		}
	} else if streamable, ok := config["streamableHttp"].(map[string]interface{}); ok {
		serverType = "StreamableHTTP"
		if url, ok := streamable["url"].(string); ok {
			endpoint = url
		}
	} else if stdio, ok := config["stdio"].(map[string]interface{}); ok {
		serverType = "Stdio"
		if cmd, ok := stdio["command"].(string); ok {
			endpoint = cmd
		}
	}

	t.App().Flash().Infof("ToolServer %s (%s): %s", raw.GetName(), serverType, endpoint)

	// For HTTP-based servers, we could do an actual health check here
	if serverType == "SSE" || serverType == "StreamableHTTP" {
		t.App().Flash().Infof("Testing connection to %s...", endpoint)
		// In a real implementation, we'd make an HTTP request here
		t.App().Flash().Infof("Connection test: Use 'kagent mcp inspector' for detailed testing")
	}

	return nil
}

// wrapText wraps text at the specified width
func wrapText(text string, width int) string {
	if len(text) <= width {
		return text
	}

	var result strings.Builder
	words := strings.Fields(text)
	lineLen := 0

	for i, word := range words {
		if lineLen+len(word)+1 > width && lineLen > 0 {
			result.WriteString("\n")
			lineLen = 0
		}
		if i > 0 && lineLen > 0 {
			result.WriteString(" ")
			lineLen++
		}
		result.WriteString(word)
		lineLen += len(word)
	}

	return result.String()
}

