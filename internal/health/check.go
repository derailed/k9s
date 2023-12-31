// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package health

import (
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// Check tracks resource health.
type Check struct {
	Counts

	GVR string
}

// Checks represents a collection of health checks.
type Checks []*Check

// NewCheck returns a new health check.
func NewCheck(gvr string) *Check {
	return &Check{
		GVR:    gvr,
		Counts: make(Counts),
	}
}

// Set sets a health metric.
func (c *Check) Set(l Level, v int64) {
	c.Counts[l] = v
}

// Inc increments a health metric.
func (c *Check) Inc(l Level) {
	c.Counts[l]++
}

// Total stores a metric total.
func (c *Check) Total(n int64) {
	c.Counts[Corpus] = n
}

// Tally retrieves a given health metric.
func (c *Check) Tally(l Level) int64 {
	return c.Counts[l]
}

// GetObjectKind returns a schema object.
func (Check) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c Check) DeepCopyObject() runtime.Object {
	return c
}
