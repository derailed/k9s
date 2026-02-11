// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultKToolServerHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "TYPE"},
	model1.HeaderColumn{Name: "URL/CMD"},
	model1.HeaderColumn{Name: "TOOLS", Attrs: model1.Attrs{Align: tview.AlignRight}},
	model1.HeaderColumn{Name: "STATUS"},
	model1.HeaderColumn{Name: "DESCRIPTION", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// KToolServer renders a kagent ToolServer to screen.
type KToolServer struct {
	Base
}

// ColorerFunc colors a resource row.
func (KToolServer) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)
		if c == model1.ErrColor {
			return c
		}

		// Color by server type
		typeIdx, ok := h.IndexOf("TYPE", true)
		if ok && typeIdx < len(re.Row.Fields) {
			switch strings.ToLower(re.Row.Fields[typeIdx]) {
			case "stdio":
				return tcell.ColorMediumAquamarine
			case "sse":
				return tcell.ColorDodgerBlue
			case "streamablehttp":
				return tcell.ColorMediumOrchid
			}
		}

		// Color by tools count
		toolsIdx, ok := h.IndexOf("TOOLS", true)
		if ok && toolsIdx < len(re.Row.Fields) {
			if re.Row.Fields[toolsIdx] == "0" {
				return tcell.ColorGray
			}
		}

		return c
	}
}

// Header returns a header row.
func (t KToolServer) Header(_ string) model1.Header {
	return t.doHeader(defaultKToolServerHeader)
}

// Render renders a K8s resource to screen.
func (t KToolServer) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := t.defaultRow(raw, row); err != nil {
		return err
	}
	if t.specs.isEmpty() {
		return nil
	}
	cols, err := t.specs.realize(raw, defaultKToolServerHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (KToolServer) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	r.ID = client.FQN(raw.GetNamespace(), raw.GetName())

	// Extract spec fields
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	description := extractString(spec, "description")

	// Extract config to determine type
	config, _ := spec["config"].(map[string]interface{})
	serverType := extractString(config, "type")
	urlOrCmd := ""

	// Determine URL/command based on type
	if stdio, ok := config["stdio"].(map[string]interface{}); ok {
		serverType = "stdio"
		urlOrCmd = extractString(stdio, "command")
	} else if sse, ok := config["sse"].(map[string]interface{}); ok {
		serverType = "sse"
		urlOrCmd = extractString(sse, "url")
	} else if streamable, ok := config["streamableHttp"].(map[string]interface{}); ok {
		serverType = "streamableHttp"
		urlOrCmd = extractString(streamable, "url")
	}

	// Extract status
	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	toolsCount := "0"
	statusStr := "Unknown"

	// Count discovered tools
	if discoveredTools, ok := status["discoveredTools"].([]interface{}); ok {
		toolsCount = fmt.Sprintf("%d", len(discoveredTools))
		if len(discoveredTools) > 0 {
			statusStr = "Ready"
		} else {
			statusStr = "No Tools"
		}
	}

	// Check conditions
	if conditions, ok := status["conditions"].([]interface{}); ok {
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := cond["type"].(string); ok && t == "Ready" {
				if s, ok := cond["status"].(string); ok {
					if s == "True" {
						statusStr = "Ready"
					} else {
						statusStr = "NotReady"
					}
				}
			}
		}
	}

	r.Fields = model1.Fields{
		raw.GetNamespace(),
		raw.GetName(),
		serverType,
		truncateStr(urlOrCmd, 40),
		toolsCount,
		statusStr,
		truncateStr(description, 50),
		mapToStr(raw.GetLabels()),
		AsStatus(diagnoseToolServer(statusStr, toolsCount)),
		ToAge(raw.GetCreationTimestamp()),
	}

	return nil
}

// Healthy checks component health.
func (t KToolServer) Healthy(_ context.Context, o any) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	statusStr := "Unknown"
	toolsCount := "0"

	if discoveredTools, ok := status["discoveredTools"].([]interface{}); ok {
		toolsCount = fmt.Sprintf("%d", len(discoveredTools))
	}

	if conditions, ok := status["conditions"].([]interface{}); ok {
		for _, c := range conditions {
			cond, ok := c.(map[string]interface{})
			if !ok {
				continue
			}
			if t, ok := cond["type"].(string); ok && t == "Ready" {
				if s, ok := cond["status"].(string); ok {
					if s == "True" {
						statusStr = "Ready"
					} else {
						statusStr = "NotReady"
					}
				}
			}
		}
	}

	return diagnoseToolServer(statusStr, toolsCount)
}

func diagnoseToolServer(status, toolsCount string) error {
	if status != "Ready" {
		return fmt.Errorf("tool server not ready: %s", status)
	}
	if toolsCount == "0" {
		return fmt.Errorf("no tools discovered")
	}
	return nil
}
