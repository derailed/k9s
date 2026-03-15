//go:build integration
// +build integration

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"testing"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

// TestCosLimitsWithPodResources_Integration tests cosLimits with pod-level resources
func TestCosLimitsWithPodResources_Integration(t *testing.T) {
	// Create containers with limits (should be ignored when pod resources exist)
	cc := []v1.Container{
		{
			Name: "c1",
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("20m"),
					v1.ResourceMemory: resource.MustParse("2Mi"),
				},
			},
		},
		{
			Name: "c2",
			Resources: v1.ResourceRequirements{
				Limits: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("40m"),
					v1.ResourceMemory: resource.MustParse("4Mi"),
				},
			},
		},
	}

	// Pod-level resources (these should take priority)
	resources := &v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("500m"),
			v1.ResourceMemory: resource.MustParse("512Mi"),
		},
	}

	// Act: Get limits using pod-level resources
	c, m, g := cosLimits(cc, resources)

	// Assert: Pod-level resources are used (not container sums)
	assertEqual(t, "500m", c.String(), "CPU should use pod-level 500m")
	assertEqual(t, "512Mi", m.String(), "Memory should use pod-level 512Mi")
	assertTrue(t, g.IsZero(), "GPU should be zero")
}

// TestCosRequestsWithPodResources_Integration tests cosRequests with pod-level resources
func TestCosRequestsWithPodResources_Integration(t *testing.T) {
	// Create containers with requests (should be ignored when pod resources exist)
	cc := []v1.Container{
		{
			Name: "c1",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("1Mi"),
				},
			},
		},
		{
			Name: "c2",
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceCPU:    resource.MustParse("10m"),
					v1.ResourceMemory: resource.MustParse("1Mi"),
				},
			},
		},
	}

	// Pod-level resources (these should take priority)
	resources := &v1.ResourceRequirements{
		Requests: v1.ResourceList{
			v1.ResourceCPU:    resource.MustParse("100m"),
			v1.ResourceMemory: resource.MustParse("128Mi"),
		},
	}

	// Act: Get requests using pod-level resources
	c, m, g := cosRequests(cc, resources)

	// Assert: Pod-level resources are used (not container sums)
	assertEqual(t, "100m", c.String(), "CPU should use pod-level 100m")
	assertEqual(t, "128Mi", m.String(), "Memory should use pod-level 128Mi")
	assertTrue(t, g.IsZero(), "GPU should be zero")
}

// Helper functions
func assertEqual(t *testing.T, expected, actual, msg string) {
	t.Helper()
	if expected != actual {
		t.Errorf("%s: expected %s, got %s", msg, expected, actual)
	}
}

func assertTrue(t *testing.T, condition bool, msg string) {
	t.Helper()
	if !condition {
		t.Errorf("%s: expected true, got false", msg)
	}
}
