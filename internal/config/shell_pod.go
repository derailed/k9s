// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"github.com/derailed/k9s/internal/client"
	v1 "k8s.io/api/core/v1"
)

const defaultDockerShellImage = "busybox:1.35.0"

// Limits represents resource limits.
type Limits map[v1.ResourceName]string

// ShellPod represents k9s shell configuration.
type ShellPod struct {
	Image            string                    `yaml:"image"`
	Command          []string                  `yaml:"command,omitempty"`
	Args             []string                  `yaml:"args,omitempty"`
	Namespace        string                    `yaml:"namespace"`
	Limits           Limits                    `yaml:"limits,omitempty"`
	Labels           map[string]string         `yaml:"labels,omitempty"`
	ImagePullSecrets []v1.LocalObjectReference `yaml:"imagePullSecrets,omitempty"`
	ImagePullPolicy  v1.PullPolicy             `yaml:"imagePullPolicy,omitempty"`
	TTY              bool                      `yaml:"tty,omitempty"`
}

// NewShellPod returns a new instance.
func NewShellPod() *ShellPod {
	return &ShellPod{
		Image:     defaultDockerShellImage,
		Namespace: "default",
		Limits:    defaultLimits(),
	}
}

// Validate validates the configuration.
func (s *ShellPod) Validate(client.Connection, KubeSettings) {
	if s.Image == "" {
		s.Image = defaultDockerShellImage
	}
	if len(s.Limits) == 0 {
		s.Limits = defaultLimits()
	}
}

func defaultLimits() Limits {
	return Limits{
		v1.ResourceCPU:    "100m",
		v1.ResourceMemory: "100Mi",
	}
}
