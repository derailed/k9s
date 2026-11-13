// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

func newStr(s string) *string {
	return &s
}

func TestComputeShellArgs(t *testing.T) {
	uu := map[string]struct {
		fqn, co, os string
		cfg         *genericclioptions.ConfigFlags
		e           string
	}{
		"config": {
			fqn: "fred/blee",
			co:  "c1",
			os:  "darwin",
			cfg: &genericclioptions.ConfigFlags{
				KubeConfig: newStr("coolConfig"),
			},
			e: "exec -it -n fred blee --kubeconfig coolConfig -c c1 -- sh -c " + shellCheck,
		},

		"no-config": {
			fqn: "fred/blee",
			co:  "c1",
			os:  "linux",
			e:   "exec -it -n fred blee -c c1 -- sh -c " + shellCheck,
		},

		"empty-config": {
			fqn: "fred/blee",
			cfg: new(genericclioptions.ConfigFlags),
			e:   "exec -it -n fred blee -- sh -c " + shellCheck,
		},

		"single-container": {
			fqn: "fred/blee",
			os:  "linux",
			cfg: new(genericclioptions.ConfigFlags),
			e:   "exec -it -n fred blee -- sh -c " + shellCheck,
		},

		"windows": {
			fqn: "fred/blee",
			co:  "c1",
			os:  windowsOS,
			cfg: new(genericclioptions.ConfigFlags),
			e:   "exec -it -n fred blee -c c1 -- cmd /c " + winShellCheck,
		},

		"full": {
			fqn: "fred/blee",
			co:  "c1",
			os:  windowsOS,
			cfg: &genericclioptions.ConfigFlags{
				KubeConfig:  newStr("coolConfig"),
				Context:     newStr("coolContext"),
				BearerToken: newStr("coolToken"),
			},
			e: "exec -it -n fred blee --kubeconfig coolConfig --context coolContext --token coolToken -c c1 -- cmd /c " + winShellCheck,
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			args := computeShellArgs(u.fqn, u.co, u.cfg, u.os)
			assert.Equal(t, u.e, strings.Join(args, " "))
		})
	}
}

func TestIsPodRunning(t *testing.T) {
	running := v1.ContainerStatus{
		Name:  "c1",
		Ready: true,
		State: v1.ContainerState{Running: &v1.ContainerStateRunning{}},
	}
	completed := v1.ContainerStatus{
		Name: "c2",
		State: v1.ContainerState{
			Terminated: &v1.ContainerStateTerminated{Reason: "Completed"},
		},
	}
	readyCond := v1.PodCondition{Type: v1.PodReady, Status: v1.ConditionTrue}

	uu := map[string]struct {
		st v1.PodStatus
		e  bool
	}{
		"running": {
			st: v1.PodStatus{
				Phase:             v1.PodRunning,
				Conditions:        []v1.PodCondition{readyCond},
				ContainerStatuses: []v1.ContainerStatus{running},
			},
			e: true,
		},
		"completed-and-running-ready": {
			st: v1.PodStatus{
				Phase:             v1.PodRunning,
				Conditions:        []v1.PodCondition{readyCond},
				ContainerStatuses: []v1.ContainerStatus{running, completed},
			},
			e: true,
		},
		"completed-and-running-not-ready": {
			st: v1.PodStatus{
				Phase:             v1.PodRunning,
				ContainerStatuses: []v1.ContainerStatus{running, completed},
			},
			e: true,
		},
		"pending": {
			st: v1.PodStatus{Phase: v1.PodPending},
		},
		"succeeded": {
			st: v1.PodStatus{
				Phase:             v1.PodSucceeded,
				ContainerStatuses: []v1.ContainerStatus{completed},
			},
		},
	}

	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			assert.Equal(t, u.e, isPodRunning(&v1.Pod{Status: u.st}))
		})
	}
}
