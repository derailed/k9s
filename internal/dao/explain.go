// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ Accessor = (*Explain)(nil)

// Explain represents a kubectl explain accessor.
type Explain struct {
	NonResource
}

// FieldInfo represents information about a field in the explain output.
type FieldInfo struct {
	Name       string
	Type       string
	Required   bool
	Depth      int // Indentation level
	EnumValues string
}

// ExplainResult represents the result of a kubectl explain command.
type ExplainResult struct {
	Path        string
	Content     string
	Kind        string
	Version     string
	Fields      []string
	FieldTypes  map[string]string // Map of field name to type
	FieldTree   []FieldInfo       // Recursive tree of all fields
	IsLeaf      bool
	LeafField   string // For leaf nodes: the field name
	LeafType    string // For leaf nodes: the field type
	Description string
}

// List returns a collection of explain results.
func (*Explain) List(_ context.Context, _ string) ([]runtime.Object, error) {
	return nil, fmt.Errorf("list not supported for explain")
}

// Get returns a kubectl explain result for a given path.
func (*Explain) Get(_ context.Context, _ string) (runtime.Object, error) {
	return nil, fmt.Errorf("get not supported for explain")
}

// Explain executes kubectl explain for the given resource path.
func (e *Explain) Explain(ctx context.Context, resourcePath string) (*ExplainResult, error) {
	return e.explainWithOptions(ctx, resourcePath, false)
}

// ExplainRecursive executes kubectl explain with --recursive flag.
func (e *Explain) ExplainRecursive(ctx context.Context, resourcePath string) (*ExplainResult, error) {
	return e.explainWithOptions(ctx, resourcePath, true)
}

// explainWithOptions executes kubectl explain with optional --recursive flag.
func (*Explain) explainWithOptions(ctx context.Context, resourcePath string, recursive bool) (*ExplainResult, error) {
	factory, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok || factory == nil {
		return nil, fmt.Errorf("no factory found in context")
	}

	// Build kubectl explain command
	args := []string{"explain", resourcePath}
	if recursive {
		args = append(args, "--recursive")
	}

	// Get kubeconfig path if available
	if kubeconfig := factory.Client().Config().Flags().KubeConfig; kubeconfig != nil && *kubeconfig != "" {
		args = append(args, "--kubeconfig", *kubeconfig)
	}

	// Get current context
	if contextName, err := factory.Client().Config().CurrentContextName(); err == nil && contextName != "" {
		args = append(args, "--context", contextName)
	}

	cmd := exec.CommandContext(ctx, "kubectl", args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("kubectl explain failed: %w: %s", err, string(output))
	}

	result := &ExplainResult{
		Path:       resourcePath,
		Content:    string(output),
		Fields:     []string{},
		FieldTypes: make(map[string]string),
		FieldTree:  []FieldInfo{},
		IsLeaf:     false,
	}

	// Parse the output to extract fields and metadata
	if recursive {
		result.parseRecursiveTree()
	} else {
		result.parseOutput()
	}

	return result, nil
}

// parseOutput parses kubectl explain output to extract field information.
func (r *ExplainResult) parseOutput() {
	lines := strings.Split(r.Content, "\n")
	inFieldsSection := false
	inDescription := false
	descriptionLines := []string{}

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Extract KIND
		if strings.HasPrefix(trimmed, "KIND:") {
			r.Kind = strings.TrimSpace(strings.TrimPrefix(trimmed, "KIND:"))
			continue
		}

		// Extract VERSION
		if strings.HasPrefix(trimmed, "VERSION:") {
			r.Version = strings.TrimSpace(strings.TrimPrefix(trimmed, "VERSION:"))
			continue
		}

		// Extract FIELD (for leaf nodes) - format: "FIELD: fieldName <type>"
		if strings.HasPrefix(trimmed, "FIELD:") {
			fieldLine := strings.TrimSpace(strings.TrimPrefix(trimmed, "FIELD:"))
			// Parse "fieldName <type>"
			if strings.Contains(fieldLine, "<") {
				parts := strings.SplitN(fieldLine, "<", 2)
				if len(parts) == 2 {
					r.LeafField = strings.TrimSpace(parts[0])
					r.LeafType = "<" + strings.TrimSpace(parts[1])
				}
			}
			r.IsLeaf = true
			continue
		}

		// Check if we're in the FIELDS section
		if strings.HasPrefix(trimmed, "FIELDS:") {
			inFieldsSection = true
			inDescription = false
			continue
		}

		// Extract description
		if strings.HasPrefix(trimmed, "DESCRIPTION:") {
			inDescription = true
			inFieldsSection = false
			continue
		}

		// Collect description lines
		if inDescription && trimmed != "" && !strings.HasPrefix(trimmed, "FIELD:") {
			descriptionLines = append(descriptionLines, trimmed)
		}

		// Parse field names - they start with exactly 2 spaces and contain a tab
		if inFieldsSection && strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") {
			// Field lines look like: "   fieldName\t<type>"
			// Split on tab to get field name and type
			if strings.Contains(line, "\t") {
				parts := strings.Split(strings.TrimSpace(line), "\t")
				if len(parts) > 0 && parts[0] != "" {
					fieldName := strings.TrimSpace(parts[0])
					r.Fields = append(r.Fields, fieldName)

					// Extract type (e.g., "<string>", "<PodSpec>", "<[]Object>")
					if len(parts) > 1 {
						fieldType := strings.TrimSpace(parts[1])
						r.FieldTypes[fieldName] = fieldType
					}
				}
			}
		}
	}

	// If no fields were found, this is likely a leaf node
	r.IsLeaf = len(r.Fields) == 0

	// Store the description
	if len(descriptionLines) > 0 {
		r.Description = strings.Join(descriptionLines, " ")
	}
}

// parseRecursiveTree parses the recursive kubectl explain output to build a complete field tree.
func (r *ExplainResult) parseRecursiveTree() {
	lines := strings.Split(r.Content, "\n")
	inFieldsSection := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Check if we're in the FIELDS section
		if strings.HasPrefix(trimmed, "FIELDS:") {
			inFieldsSection = true
			continue
		}

		// Skip non-field lines
		if !inFieldsSection || trimmed == "" {
			continue
		}

		// Count leading spaces to determine depth
		// Each level is indented by 2 spaces
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		if leadingSpaces < 2 {
			continue
		}

		depth := leadingSpaces / 2

		// Parse field line: "  fieldName\t<type> -required-"
		if strings.Contains(line, "\t") {
			parts := strings.Split(strings.TrimSpace(line), "\t")
			if len(parts) == 0 || parts[0] == "" {
				continue
			}

			fieldName := strings.TrimSpace(parts[0])
			fieldType := ""
			required := false
			enumValues := ""

			// Parse type and flags
			if len(parts) > 1 {
				typeInfo := strings.TrimSpace(parts[1])

				// Check for -required- flag
				if strings.Contains(typeInfo, "-required-") {
					required = true
					typeInfo = strings.ReplaceAll(typeInfo, "-required-", "")
					typeInfo = strings.TrimSpace(typeInfo)
				}

				fieldType = typeInfo
			}

			// Check next line for enum values
			// (enum values appear on the next line after "enum:")
			if i+1 < len(lines) {
				nextLine := strings.TrimSpace(lines[i+1])
				if strings.HasPrefix(nextLine, "enum:") {
					enumValues = strings.TrimPrefix(nextLine, "enum:")
					enumValues = strings.TrimSpace(enumValues)
				}
			}

			r.FieldTree = append(r.FieldTree, FieldInfo{
				Name:       fieldName,
				Type:       fieldType,
				Required:   required,
				EnumValues: enumValues,
				Depth:      depth - 1, // Adjust depth (first level is 2 spaces = depth 0)
			})
		}
	}
}

// GetFields returns the list of available fields for a resource path.
func (e *Explain) GetFields(ctx context.Context, resourcePath string) ([]string, error) {
	result, err := e.Explain(ctx, resourcePath)
	if err != nil {
		return nil, err
	}

	return result.Fields, nil
}
