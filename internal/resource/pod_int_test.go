package resource

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
)

func TestPodStatuses(t *testing.T) {
	type counts struct {
		ready, terminated, restarts int
	}

	uu := []struct {
		s []v1.ContainerStatus
		e counts
	}{
		{
			[]v1.ContainerStatus{
				v1.ContainerStatus{
					Name:  "c1",
					Ready: true,
					State: v1.ContainerState{
						Running: &v1.ContainerStateRunning{},
					},
				},
				v1.ContainerStatus{
					Name:         "c2",
					Ready:        false,
					RestartCount: 10,
					State: v1.ContainerState{
						Terminated: &v1.ContainerStateTerminated{},
					},
				},
			},
			counts{1, 1, 10},
		},
	}

	var p Pod
	for _, u := range uu {
		cr, ct, cs := p.statuses(u.s)
		assert.Equal(t, u.e.ready, cr)
		assert.Equal(t, u.e.terminated, ct)
		assert.Equal(t, u.e.restarts, cs)
	}
}

func TestPodPhase(t *testing.T) {
	uu := []struct {
		s v1.PodStatus
		e string
	}{
		{
			v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Running: &v1.ContainerStateRunning{},
						},
					},
				},
			},
			"Running",
		},
		{
			v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Waiting: &v1.ContainerStateWaiting{
								Reason: "blee",
							},
						},
					},
				},
			},
			"blee",
		},
		{
			v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{},
						},
					},
				},
			},
			"Terminating",
		},
		{
			v1.PodStatus{
				ContainerStatuses: []v1.ContainerStatus{
					{
						Name: "c1",
						State: v1.ContainerState{
							Terminated: &v1.ContainerStateTerminated{
								Reason: "blee",
							},
						},
					},
				},
			},
			"blee",
		},
	}

	var p Pod
	for _, u := range uu {
		assert.Equal(t, u.e, p.phase(u.s))
	}
}
