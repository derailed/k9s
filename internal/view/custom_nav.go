// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"strings"
	"text/template"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
)

// customNavigate handles custom navigation between CRDs based on user configuration.
func customNavigate(app *App, sourceGVR *client.GVR, sourcePath string, rule *config.NavigationRule) error {
	// Get the source resource to extract metadata for templating
	sourceObj, err := fetchResourceObject(app, sourceGVR, sourcePath)
	if err != nil {
		return fmt.Errorf("failed to fetch source resource: %w", err)
	}

	// Parse target GVR
	targetGVR := client.NewGVR(rule.TargetGVR)

	// Determine target namespace
	targetNS, err := determineTargetNamespace(sourcePath, rule.TargetNamespace, sourceObj)
	if err != nil {
		return err
	}

	// Create new view for target resource
	v := NewBrowser(targetGVR)

	// Build label selector if specified
	var labelSel labels.Selector
	if rule.LabelSelector != "" {
		labelStr, labelErr := applyTemplate(rule.LabelSelector, sourceObj)
		if labelErr != nil {
			return fmt.Errorf("failed to process label selector template: %w", labelErr)
		}
		labelSel, labelErr = labels.Parse(labelStr)
		if labelErr != nil {
			return fmt.Errorf("invalid label selector: %w", labelErr)
		}
	}

	// Build field selector if specified
	var fieldSel string
	if rule.FieldSelector != "" {
		var fieldErr error
		fieldSel, fieldErr = applyTemplate(rule.FieldSelector, sourceObj)
		if fieldErr != nil {
			return fmt.Errorf("failed to process field selector template: %w", fieldErr)
		}
	}

	// Set up the context for the navigation
	v.SetContextFn(func(ctx context.Context) context.Context {
		if fieldSel != "" {
			ctx = context.WithValue(ctx, internal.KeyFields, fieldSel)
		}
		return ctx
	})

	// Set label selector if we have one
	if labelSel != nil {
		v.SetLabelSelector(labelSel, true)
	}

	// Navigate to the target namespace if needed
	if targetNS != "" && targetNS != client.NamespaceAll {
		if err := app.Config.SetActiveNamespace(targetNS); err != nil {
			slog.Error("Unable to set active namespace during custom navigation", slogs.Error, err)
		}
	}

	// Inject the new view
	if err := app.inject(v, false); err != nil {
		return fmt.Errorf("failed to inject target view: %w", err)
	}

	return nil
}

// fetchResourceObject retrieves the full resource object for templating.
func fetchResourceObject(app *App, gvr *client.GVR, path string) (map[string]any, error) {
	accessor, err := dao.AccessorFor(app.factory, gvr)
	if err != nil {
		return nil, err
	}

	o, err := accessor.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	u, ok := o.(*unstructured.Unstructured)
	if !ok {
		return nil, fmt.Errorf("unexpected object type: %T", o)
	}

	return u.Object, nil
}

// determineTargetNamespace determines which namespace to use for the target resource.
func determineTargetNamespace(sourcePath, targetNSConfig string, sourceObj map[string]any) (string, error) {
	sourceNS, _ := client.Namespaced(sourcePath)

	switch targetNSConfig {
	case "", "same":
		return sourceNS, nil
	case "all":
		return client.NamespaceAll, nil
	default:
		// Check if it contains template syntax
		if strings.Contains(targetNSConfig, "{{") {
			return applyTemplate(targetNSConfig, sourceObj)
		}
		// It's a specific namespace name
		return targetNSConfig, nil
	}
}

// applyTemplate applies Go template to the selector using the source resource data.
func applyTemplate(tmplStr string, data map[string]any) (string, error) {
	tmpl, err := template.New("selector").Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}
