// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
)

func TestApplyContainerOverrides(t *testing.T) {
	spec := v1.PodSpec{
		InitContainers: []v1.Container{{Name: "setup", Command: []string{"old-setup"}}},
		Containers: []v1.Container{
			{Name: "main", Command: []string{"old"}, Args: []string{"old-arg"}},
			{Name: "sidecar", Command: []string{"sidecar"}},
		},
	}
	overrides := map[string]ContainerOverride{
		"setup": {Command: []string{"new-setup"}, Args: []string{"setup-arg"}},
		"main": {
			Command: []string{"new", "command"},
			Args:    []string{"arg with spaces"},
			Env: []v1.EnvVar{
				{Name: "PLAIN", Value: "value"},
				{Name: "SECRET", ValueFrom: &v1.EnvVarSource{SecretKeyRef: &v1.SecretKeySelector{
					LocalObjectReference: v1.LocalObjectReference{Name: "secret"},
					Key:                  "token",
				}}},
			},
		},
	}

	require.NoError(t, applyContainerOverrides(&spec, overrides))
	assert.Equal(t, []string{"new-setup"}, spec.InitContainers[0].Command)
	assert.Equal(t, []string{"setup-arg"}, spec.InitContainers[0].Args)
	assert.Equal(t, []string{"new", "command"}, spec.Containers[0].Command)
	assert.Equal(t, []string{"arg with spaces"}, spec.Containers[0].Args)
	assert.Equal(t, "value", spec.Containers[0].Env[0].Value)
	assert.Equal(t, "secret", spec.Containers[0].Env[1].ValueFrom.SecretKeyRef.Name)
	assert.Equal(t, []string{"sidecar"}, spec.Containers[1].Command)

	overrides["main"].Command[0] = "changed"
	overrides["main"].Env[1].ValueFrom.SecretKeyRef.Name = "changed"
	assert.Equal(t, "new", spec.Containers[0].Command[0])
	assert.Equal(t, "secret", spec.Containers[0].Env[1].ValueFrom.SecretKeyRef.Name)
}

func TestApplyContainerOverridesMissingContainer(t *testing.T) {
	err := applyContainerOverrides(&v1.PodSpec{}, map[string]ContainerOverride{"missing": {}})
	require.EqualError(t, err, `container "missing" not found in CronJob`)
}
