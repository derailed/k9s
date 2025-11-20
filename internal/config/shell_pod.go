// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	v1 "k8s.io/api/core/v1"
)

const defaultDockerShellImage = "busybox:1.37.0"

// Limits represents resource limits.
type Limits map[v1.ResourceName]string

// ShellPod represents k9s shell configuration.
type ShellPod struct {
	Image            string                    `json:"image" yaml:"image"`
	Command          []string                  `json:"command,omitempty" yaml:"command,omitempty"`
	Args             []string                  `json:"args,omitempty" yaml:"args,omitempty"`
	Namespace        string                    `json:"namespace" yaml:"namespace"`
	Limits           Limits                    `json:"limits,omitempty" yaml:"limits,omitempty"`
	Labels           map[string]string         `json:"labels,omitempty" yaml:"labels,omitempty"`
	ImagePullSecrets []v1.LocalObjectReference `json:"imagePullSecrets,omitempty" yaml:"imagePullSecrets,omitempty"`
	ImagePullPolicy  v1.PullPolicy             `json:"imagePullPolicy,omitempty" yaml:"imagePullPolicy,omitempty"`
	TTY              bool                      `json:"tty,omitempty" yaml:"tty,omitempty"`
	HostPathVolume   []hostPathVolume          `json:"hostPathVolume,omitempty" yaml:"hostPathVolume,omitempty"`
}

type hostPathVolume struct {
	Name      string `json:"name" yaml:"name"`
	MountPath string `json:"mountPath" yaml:"mountPath"`
	HostPath  string `json:"hostPath" yaml:"hostPath"`
	ReadOnly  bool   `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
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
func (s *ShellPod) Validate() {
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
