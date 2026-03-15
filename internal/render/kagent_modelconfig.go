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

var defaultKModelConfigHeader = model1.Header{
	model1.HeaderColumn{Name: "NAMESPACE"},
	model1.HeaderColumn{Name: "NAME"},
	model1.HeaderColumn{Name: "PROVIDER"},
	model1.HeaderColumn{Name: "MODEL"},
	model1.HeaderColumn{Name: "SECRET"},
	model1.HeaderColumn{Name: "ACCEPTED"},
	model1.HeaderColumn{Name: "TLS"},
	model1.HeaderColumn{Name: "LABELS", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "VALID", Attrs: model1.Attrs{Wide: true}},
	model1.HeaderColumn{Name: "AGE", Attrs: model1.Attrs{Time: true}},
}

// KModelConfig renders a kagent ModelConfig to screen.
type KModelConfig struct {
	Base
}

// ColorerFunc colors a resource row.
func (KModelConfig) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, re *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, re)
		if c == model1.ErrColor {
			return c
		}

		acceptedIdx, ok := h.IndexOf("ACCEPTED", true)
		if ok && acceptedIdx < len(re.Row.Fields) {
			switch strings.ToLower(re.Row.Fields[acceptedIdx]) {
			case "true":
				return tcell.ColorMediumSpringGreen
			case "false":
				return tcell.ColorOrangeRed
			}
		}

		// Color by provider type
		providerIdx, ok := h.IndexOf("PROVIDER", true)
		if ok && providerIdx < len(re.Row.Fields) {
			switch re.Row.Fields[providerIdx] {
			case "OpenAI":
				return tcell.ColorDodgerBlue
			case "Anthropic":
				return tcell.ColorSandyBrown
			case "AzureOpenAI":
				return tcell.ColorCornflowerBlue
			case "Ollama":
				return tcell.ColorMediumPurple
			case "Gemini", "GeminiVertexAI":
				return tcell.ColorGold
			}
		}

		return c
	}
}

// Header returns a header row.
func (m KModelConfig) Header(_ string) model1.Header {
	return m.doHeader(defaultKModelConfigHeader)
}

// Render renders a K8s resource to screen.
func (m KModelConfig) Render(o any, _ string, row *model1.Row) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return fmt.Errorf("expected Unstructured, but got %T", o)
	}
	if err := m.defaultRow(raw, row); err != nil {
		return err
	}
	if m.specs.isEmpty() {
		return nil
	}
	cols, err := m.specs.realize(raw, defaultKModelConfigHeader, row)
	if err != nil {
		return err
	}
	cols.hydrateRow(row)

	return nil
}

func (KModelConfig) defaultRow(raw *unstructured.Unstructured, r *model1.Row) error {
	r.ID = client.FQN(raw.GetNamespace(), raw.GetName())

	// Extract spec fields
	spec, _, _ := unstructured.NestedMap(raw.Object, "spec")
	provider := extractString(spec, "provider")
	model := extractString(spec, "model")
	apiKeySecret := extractString(spec, "apiKeySecret")

	// Check TLS config
	tlsEnabled := "No"
	if tls, ok := spec["tls"].(map[string]interface{}); ok {
		if disableVerify, ok := tls["disableVerify"].(bool); ok && disableVerify {
			tlsEnabled = "Insecure"
		} else if caCert, ok := tls["caCertSecretRef"].(string); ok && caCert != "" {
			tlsEnabled = "Custom CA"
		} else {
			tlsEnabled = "Yes"
		}
	}

	// Extract status conditions
	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	accepted := extractConditionStatus(status, "Accepted")

	r.Fields = model1.Fields{
		raw.GetNamespace(),
		raw.GetName(),
		provider,
		model,
		apiKeySecret,
		accepted,
		tlsEnabled,
		mapToStr(raw.GetLabels()),
		AsStatus(diagnoseModelConfig(accepted)),
		ToAge(raw.GetCreationTimestamp()),
	}

	return nil
}

// Healthy checks component health.
func (m KModelConfig) Healthy(_ context.Context, o any) error {
	raw, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil
	}

	status, _, _ := unstructured.NestedMap(raw.Object, "status")
	accepted := extractConditionStatus(status, "Accepted")

	return diagnoseModelConfig(accepted)
}

func diagnoseModelConfig(accepted string) error {
	if strings.ToLower(accepted) != "true" {
		return errors.New("model config not accepted")
	}
	return nil
}
