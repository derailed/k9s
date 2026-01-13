// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v3"
)

// JumpRule defines a custom jump from one resource to another.
type JumpRule struct {
	// TargetGVR is the target resource to jump to (e.g., "mygroup.io/v1/patchjobs")
	TargetGVR string `yaml:"targetGVR"`

	// LabelSelector defines label-based filtering for the target resource.
	LabelSelector string `yaml:"labelSelector,omitempty"`

	// FieldSelector defines field-based filtering for the target resource.
	FieldSelector string `yaml:"fieldSelector,omitempty"`

	// TargetNamespace specifies namespace handling:
	// - "" (empty/omitted): use the source resource's namespace (default)
	// - "all": view resources across all namespaces
	// - "specific-ns": use a specific namespace name
	// Supports Go templates: "{{.metadata.namespace}}"
	TargetNamespace string `yaml:"targetNamespace,omitempty"`
}

// CustomJumps represents a collection of custom jump rules.
type CustomJumps struct {
	Jumps map[string]JumpRule `yaml:"jumps"`
}

// NewCustomJumps returns a new jumps configuration.
func NewCustomJumps() *CustomJumps {
	return &CustomJumps{
		Jumps: make(map[string]JumpRule),
	}
}

// Reset clears out configurations.
func (n *CustomJumps) Reset() {
	for k := range n.Jumps {
		delete(n.Jumps, k)
	}
}

// Load loads jump configurations.
func (n *CustomJumps) Load(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.JumpsSchema, bb); err != nil {
		slog.Warn("Jump validation failed. Please update your config and restart!",
			slogs.Path, path,
			slogs.Error, err,
		)
	}
	var in struct {
		Jumps map[string]JumpRule `yaml:"jumps"`
	}
	if err := yaml.Unmarshal(bb, &in); err != nil {
		return err
	}
	if in.Jumps != nil {
		n.Jumps = in.Jumps
	}

	return nil
}

// GetRule returns the jump rule for a given GVR, if one exists.
func (n *CustomJumps) GetRule(gvr *client.GVR) (*JumpRule, bool) {
	if rule, ok := n.Jumps[gvr.String()]; ok {
		return &rule, true
	}

	return nil, false
}

// ValidateRule validates a jump rule.
func (r *JumpRule) ValidateRule() error {
	if r.TargetGVR == "" {
		return fmt.Errorf("targetGVR is required")
	}

	if r.LabelSelector == "" && r.FieldSelector == "" {
		return fmt.Errorf("at least one of labelSelector or fieldSelector must be specified")
	}

	return nil
}
