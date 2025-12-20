// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"regexp"
	"strings"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"
	"github.com/derailed/k9s/internal/slogs"
	"gopkg.in/yaml.v3"
)

// NavigationRule defines a custom navigation from one resource to another.
type NavigationRule struct {
	// TargetGVR is the target resource to navigate to (e.g., "mygroup.io/v1/patchjobs")
	TargetGVR string `yaml:"targetGVR"`

	// LabelSelector defines label-based filtering for the target resource.
	LabelSelector string `yaml:"labelSelector,omitempty"`

	// FieldSelector defines field-based filtering for the target resource.
	FieldSelector string `yaml:"fieldSelector,omitempty"`

	// TargetNamespace specifies namespace handling:
	// - "same": use the same namespace as the source resource (default)
	// - "all": view resources across all namespaces
	// - "specific-ns": use a specific namespace name
	TargetNamespace string `yaml:"targetNamespace,omitempty"`
}

// CustomNavigations represents a collection of custom navigation rules.
type CustomNavigations struct {
	Navigations map[string]NavigationRule `yaml:"navigations"`
}

// NewCustomNavigations returns a new navigations configuration.
func NewCustomNavigations() *CustomNavigations {
	return &CustomNavigations{
		Navigations: make(map[string]NavigationRule),
	}
}

// Reset clears out configurations.
func (n *CustomNavigations) Reset() {
	for k := range n.Navigations {
		delete(n.Navigations, k)
	}
}

// Load loads navigation configurations.
func (n *CustomNavigations) Load(path string) error {
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.NavigationsSchema, bb); err != nil {
		slog.Warn("Navigation validation failed. Please update your config and restart!",
			slogs.Path, path,
			slogs.Error, err,
		)
	}
	var in struct {
		K9s *CustomNavigations `yaml:"k9s"`
	}
	if err := yaml.Unmarshal(bb, &in); err != nil {
		return err
	}
	if in.K9s != nil {
		n.Navigations = in.K9s.Navigations
	}

	return nil
}

// GetRule returns the navigation rule for a given GVR, if one exists.
// It supports exact matches and regex patterns.
func (n *CustomNavigations) GetRule(gvr string) (*NavigationRule, bool) {
	// Try exact match first
	if rule, ok := n.Navigations[gvr]; ok {
		return &rule, true
	}

	// Try regex match
	for key, rule := range n.Navigations {
		if strings.Contains(key, "*") || strings.Contains(key, "^") || strings.Contains(key, "$") {
			if rx, err := regexp.Compile(key); err == nil && rx.MatchString(gvr) {
				return &rule, true
			}
		}
	}

	return nil, false
}

// ValidateRule validates a navigation rule.
func (r *NavigationRule) ValidateRule() error {
	if r.TargetGVR == "" {
		return fmt.Errorf("targetGVR is required")
	}

	if r.LabelSelector == "" && r.FieldSelector == "" {
		return fmt.Errorf("at least one of labelSelector or fieldSelector must be specified")
	}

	return nil
}
