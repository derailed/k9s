// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogOptions_ToKubectlCommand(t *testing.T) {
	tests := []struct {
		name     string
		opts     *LogOptions
		expected string
	}{
		{
			name: "basic pod logs",
			opts: &LogOptions{
				Path:  "default/my-pod",
				Lines: 1000,
			},
			expected: "kubectl logs -n default my-pod --tail=1000 -f",
		},
		{
			name: "pod logs with container",
			opts: &LogOptions{
				Path:      "default/my-pod",
				Container: "app",
				Lines:     500,
			},
			expected: "kubectl logs -n default my-pod -c app --tail=500 -f",
		},
		{
			name: "pod logs with all containers",
			opts: &LogOptions{
				Path:          "default/my-pod",
				AllContainers: true,
				Lines:         1000,
			},
			expected: "kubectl logs -n default my-pod --all-containers=true --tail=1000 -f",
		},
		{
			name: "previous logs",
			opts: &LogOptions{
				Path:     "default/my-pod",
				Previous: true,
				Lines:    1000,
			},
			expected: "kubectl logs -n default my-pod --previous --tail=1000 -f",
		},
		{
			name: "head mode",
			opts: &LogOptions{
				Path: "default/my-pod",
				Head: true,
			},
			expected: "kubectl logs -n default my-pod --limit-bytes=5000",
		},
		{
			name: "with since seconds (less than 1 minute)",
			opts: &LogOptions{
				Path:         "default/my-pod",
				Lines:        1000,
				SinceSeconds: 30,
			},
			expected: "kubectl logs -n default my-pod --tail=1000 --since=30s -f",
		},
		{
			name: "with since seconds (minutes)",
			opts: &LogOptions{
				Path:         "default/my-pod",
				Lines:        1000,
				SinceSeconds: 300, // 5 minutes
			},
			expected: "kubectl logs -n default my-pod --tail=1000 --since=5m -f",
		},
		{
			name: "with since seconds (hours)",
			opts: &LogOptions{
				Path:         "default/my-pod",
				Lines:        1000,
				SinceSeconds: 7200, // 2 hours
			},
			expected: "kubectl logs -n default my-pod --tail=1000 --since=2h -f",
		},
		{
			name: "with timestamps",
			opts: &LogOptions{
				Path:          "default/my-pod",
				Lines:         1000,
				ShowTimestamp: true,
			},
			expected: "kubectl logs -n default my-pod --tail=1000 -f --timestamps",
		},
		{
			name: "with since time",
			opts: &LogOptions{
				Path:      "default/my-pod",
				Lines:     1000,
				SinceTime: "2024-01-01T00:00:00Z",
			},
			expected: "kubectl logs -n default my-pod --tail=1000 --since-time=2024-01-01T00:00:00Z -f",
		},
		{
			name: "pod without namespace",
			opts: &LogOptions{
				Path:  "my-pod",
				Lines: 1000,
			},
			expected: "kubectl logs my-pod --tail=1000 -f",
		},
		{
			name: "complete example",
			opts: &LogOptions{
				Path:          "production/api-server",
				Container:     "api",
				Lines:         500,
				SinceSeconds:  300,
				ShowTimestamp: true,
			},
			expected: "kubectl logs -n production api-server -c api --tail=500 --since=5m -f --timestamps",
		},
		{
			name: "previous logs with container and timestamps",
			opts: &LogOptions{
				Path:          "default/my-pod",
				Container:     "app",
				Previous:      true,
				Lines:         1000,
				ShowTimestamp: true,
			},
			expected: "kubectl logs -n default my-pod -c app --previous --tail=1000 -f --timestamps",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.ToKubectlCommand()
			assert.Equal(t, tt.expected, result)
		})
	}
}
