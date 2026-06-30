// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"context"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/xray"
)

// TreeModel represents a generic tree data model used by the Xray view.
type TreeModel interface {
	// SetRefreshRate sets the model refresh rate.
	SetRefreshRate(time.Duration)

	// SetNamespace sets the active namespace.
	SetNamespace(string)

	// GetNamespace returns the current namespace.
	GetNamespace() string

	// AddListener adds a tree listener.
	AddListener(TreeListener)

	// Watch starts watching for changes.
	Watch(context.Context)

	// Peek returns the current tree root.
	Peek() *xray.TreeNode

	// ClearFilter clears the current filter.
	ClearFilter()

	// SetFilter sets the current filter.
	SetFilter(string)

	// ToYAML returns the YAML representation of a resource.
	ToYAML(ctx context.Context, gvr *client.GVR, path string) (string, error)

	// Describe returns a text description of a resource.
	Describe(ctx context.Context, gvr *client.GVR, path string) (string, error)
}
