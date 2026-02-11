// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

var defaultKAgentHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "TYPE"},
	model1.HeaderColumn{Name: "MODEL"},
	model1.HeaderColumn{Name: "TOOLS", Attrs: model1.Attrs{Align: 1}},
	model1.HeaderColumn{Name: "READY"},
	model1.HeaderColumn{Name: "ACCEPTED"},
	model1.HeaderColumn{Name: "DESCRIPTION", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// KAgent renders a kagent Agent to screen.
type KAgent struct {
	Base
}

// ColorerFunc colors a resource row.
func (KAgent) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)
		if c == model1.ErrColor {
			return c
		}

		readyIdx, ok := h.IndexOf("READY", true)
		if ok && readyIdx < len(re.Row.Fields) {
			switch strings.ToLower(re.Row.Fields[readyIdx]) {
			case "true":
				return tcell.ColorMediumSpringGreen
			case "false":
				return tcell.ColorOrangeRed
			}
		}

		return c
	}
}

// Header returns a header row.
func (a KAgent) Header(_ string) model1.Header {
	return a.doHeader(defaultKAgentHeader)
}

// Render renders a K8s resource to screen.
func (a KAgent) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := a.defaultRow(raw, row); err != nil {
		return err
	}
	if a.specs.isEmpty() {
		return nil
	}
	cols, err := a.specs.realize(raw, defaultKAgentHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (KAgent) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	r.ID = client.FQN(raw.GetNamespace(), raw.GetName())

	// Extract spec fields
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	agentType := extractString(spec, "type")
	description := extractString(spec, "description")

	// Extract model config and tools count from declarative spec
	modelConfig := NAValue
	toolsCount := "0"
	if declarative, ok := spec["declarative"].(map[string]interface{}); ok {
		if mc, ok := declarative["modelConfig"].(string); ok && mc != "" {
			modelConfig = mc
		}
		if tools, ok := declarative["tools"].([]interface{}); ok {
			toolsCount = fmt.Sprintf("%d", len(tools))
		}
	}

	// Extract status conditions
	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	ready := extractConditionStatus(status, "Ready")
	accepted := extractConditionStatus(status, "Accepted")

	r.Fields = model1.Fields{
		raw.GetNamespace(),
		raw.GetName(),
		agentType,
		modelConfig,
		toolsCount,
		ready,
		accepted,
		truncateStr(description, 50),
		mapToStr(raw.GetLabels()),
		AsStatus(diagnoseAgent(ready, accepted)),
		ToAge(raw.GetCreationTimestamp()),
	}

	return nil
}

// Healthy checks component health.
func (a KAgent) Healthy(_ context.Context, o any) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	ready := extractConditionStatus(status, "Ready")
	accepted := extractConditionStatus(status, "Accepted")

	return diagnoseAgent(ready, accepted)
}

func diagnoseAgent(ready, accepted string) error {
	if strings.ToLower(ready) != "true" {
		return errors.New("agent not ready")
	}
	if strings.ToLower(accepted) != "true" {
		return errors.New("agent not accepted")
	}
	return nil
}

// extractConditionStatus extracts a condition status from status.conditions
func extractConditionStatus(status map[string]interface{}, conditionType string) string {
	conditions, ok := status["conditions"].([]interface{})
	if !ok {
		return NAValue
	}

	for _, c := range conditions {
		cond, ok := c.(map[string]interface{})
		if !ok {
			continue
		}
		if t, ok := cond["type"].(string); ok && t == conditionType {
			if s, ok := cond["status"].(string); ok {
				return s
			}
		}
	}
	return NAValue
}

// extractString safely extracts a string from a map
func extractString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}

// truncateStr truncates a string to maxLen characters
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
