// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// KAgent represents a kagent Agent viewer.
type KAgent struct {
	ResourceViewer
}

// NewKAgent returns a new agent viewer.
func NewKAgent(gvr *client.GVR) ResourceViewer {
	a := KAgent{
		ResourceViewer: NewBrowser(gvr),
	}
	a.AddBindKeysFn(a.bindKeys)
	a.GetTable().SetEnterFn(a.showAgentDetails)

	return &a
}

func (a *KAgent) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyI:        ui.NewKeyAction("Invoke", a.invokeCmd, true),
		ui.KeyC:        ui.NewKeyAction("Chat", a.chatCmd, true),
		ui.KeyT:        ui.NewKeyAction("Tools", a.showToolsCmd, true),
		ui.KeyM:        ui.NewKeyAction("ModelConfig", a.showModelConfigCmd, true),
		ui.KeyL:        ui.NewKeyAction("Logs", a.logsCmd, true),
		tcell.KeyCtrlI: ui.NewKeyAction("Quick Invoke", a.quickInvokeCmd, true),
	})
}

func (a *KAgent) showAgentDetails(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	// Show YAML view by default when entering an agent
	v := NewLiveView(app, yamlAction, model.NewYAML(gvr, path))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

// invokeCmd invokes the agent with a message
func (a *KAgent) invokeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)
	a.App().Flash().Infof("Invoking agent %s...", name)

	// Open external terminal with kagent invoke command
	args := []string{
		"invoke",
		"--namespace", ns,
		"--agent", name,
	}

	return a.runKAgentCommand(args)
}

// chatCmd starts an interactive chat session with the agent
func (a *KAgent) chatCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)

	// Open embedded chat view
	chat := NewKAgentChat(a.App(), ns, name)
	if err := a.App().inject(chat, false); err != nil {
		a.App().Flash().Err(err)
	}

	return nil
}

// quickInvokeCmd prompts for a message and invokes the agent
func (a *KAgent) quickInvokeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)

	// Use the command buffer to get a message
	a.App().Flash().Infof("Enter message for agent %s (press Enter to send):", name)

	// For now, just show a message - full implementation would need a dialog
	a.App().Flash().Infof("Quick invoke: kagent invoke -n %s %s --message <your-message>", ns, name)

	return nil
}

// showToolsCmd shows the tools available to this agent
func (a *KAgent) showToolsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the agent object to extract tool information
	ctx := context.Background()
	o, err := a.App().factory.Get(a.GVR(), path, true, nil)
	if err != nil {
		a.App().Flash().Errf("Failed to get agent: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		a.App().Flash().Err(fmt.Errorf("unexpected object type"))
		return nil
	}

	// Extract tools from spec.declarative.tools
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	declarative, ok := spec["declarative"].(map[string]interface{})
	if !ok {
		a.App().Flash().Warn("Agent has no declarative spec or tools")
		return nil
	}

	tools, ok := declarative["tools"].([]interface{})
	if !ok || len(tools) == 0 {
		a.App().Flash().Warn("Agent has no tools configured")
		return nil
	}

	// Build tools summary
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Tools for agent %s:\n\n", raw.GetName()))

	for i, t := range tools {
		tool, ok := t.(map[string]interface{})
		if !ok {
			continue
		}

		toolType := ""
		if tt, ok := tool["type"].(string); ok {
			toolType = tt
		}

		sb.WriteString(fmt.Sprintf("%d. Type: %s\n", i+1, toolType))

		if toolType == "McpServer" {
			if mcpServer, ok := tool["mcpServer"].(map[string]interface{}); ok {
				if name, ok := mcpServer["name"].(string); ok {
					sb.WriteString(fmt.Sprintf("   ToolServer: %s\n", name))
				}
				if toolNames, ok := mcpServer["toolNames"].([]interface{}); ok && len(toolNames) > 0 {
					names := make([]string, 0, len(toolNames))
					for _, tn := range toolNames {
						if s, ok := tn.(string); ok {
							names = append(names, s)
						}
					}
					sb.WriteString(fmt.Sprintf("   Tools: %s\n", strings.Join(names, ", ")))
				}
			}
		} else if toolType == "Agent" {
			if agent, ok := tool["agent"].(map[string]interface{}); ok {
				if name, ok := agent["name"].(string); ok {
					sb.WriteString(fmt.Sprintf("   Agent: %s\n", name))
				}
			}
		}
		sb.WriteString("\n")
	}

	// Show in a details view
	v := NewDetails(a.App(), "Agent Tools", raw.GetName(), contentTXT, true)
	v.Update(sb.String())
	if err := a.App().inject(v, false); err != nil {
		a.App().Flash().Err(err)
	}

	_ = ctx // suppress unused warning
	return nil
}

// showModelConfigCmd navigates to the associated ModelConfig
func (a *KAgent) showModelConfigCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	// Get the agent object
	o, err := a.App().factory.Get(a.GVR(), path, true, nil)
	if err != nil {
		a.App().Flash().Errf("Failed to get agent: %v", err)
		return nil
	}

	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	// Extract modelConfig from spec.declarative.modelConfig
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	declarative, ok := spec["declarative"].(map[string]interface{})
	if !ok {
		a.App().Flash().Warn("Agent has no declarative spec")
		return nil
	}

	modelConfigName, ok := declarative["modelConfig"].(string)
	if !ok || modelConfigName == "" {
		a.App().Flash().Warn("Agent has no modelConfig specified")
		return nil
	}

	// Navigate to the ModelConfig
	mcPath := fmt.Sprintf("modelconfigs %s/%s", raw.GetNamespace(), modelConfigName)
	a.App().gotoResource(mcPath, "", false, true)

	return nil
}

// logsCmd shows logs for the agent deployment
func (a *KAgent) logsCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := a.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	ns, name := client.Namespaced(path)

	// Navigate to pods with label selector for this agent
	podPath := fmt.Sprintf("pods -n %s -l kagent.dev/agent=%s", ns, name)
	a.App().gotoResource(podPath, "", false, true)

	return nil
}

// runKAgentCommand runs a kagent CLI command in an external terminal
func (a *KAgent) runKAgentCommand(args []string) *tcell.EventKey {
	// Check if kagent CLI is available
	kagentPath, err := exec.LookPath("kagent")
	if err != nil {
		a.App().Flash().Errf("kagent CLI not found in PATH. Install from: https://kagent.dev")
		return nil
	}

	// Build full command
	cmd := exec.Command(kagentPath, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Suspend the UI and run the command
	a.App().Halt()
	defer a.App().Resume()

	a.App().Suspend(func() {
		if err := cmd.Run(); err != nil {
			fmt.Printf("\nCommand failed: %v\n", err)
		}
		fmt.Println("\nPress Enter to return to k9s...")
		fmt.Scanln()
	})

	return nil
}

