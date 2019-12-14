package resource

// BOZO!!
// import (
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// 	v1 "k8s.io/api/core/v1"
// 	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
// )

// func TestPodStatuses(t *testing.T) {
// 	type counts struct {
// 		ready, terminated, restarts int
// 	}

// 	uu := []struct {
// 		s []v1.ContainerStatus
// 		e counts
// 	}{
// 		{
// 			[]v1.ContainerStatus{
// 				{
// 					Name:  "c1",
// 					Ready: true,
// 					State: v1.ContainerState{
// 						Running: &v1.ContainerStateRunning{},
// 					},
// 				},
// 				{
// 					Name:         "c2",
// 					Ready:        false,
// 					RestartCount: 10,
// 					State: v1.ContainerState{
// 						Terminated: &v1.ContainerStateTerminated{},
// 					},
// 				},
// 			},
// 			counts{1, 1, 10},
// 		},
// 	}

// 	var p Pod
// 	for _, u := range uu {
// 		cr, ct, cs := p.statuses(u.s)
// 		assert.Equal(t, u.e.ready, cr)
// 		assert.Equal(t, u.e.terminated, ct)
// 		assert.Equal(t, u.e.restarts, cs)
// 	}
// }

// func TestPodPhase(t *testing.T) {
// 	uu := []struct {
// 		p *v1.Pod
// 		e string
// 	}{
// 		{makePodStatus("p1", v1.PodRunning, ""), "Running"},
// 		{makePodStatus("p2", v1.PodRunning, "Evicted"), "Evicted"},
// 		{makePodStatus("p1", v1.PodPending, ""), "Pending"},
// 		{makePodStatus("p1", v1.PodSucceeded, ""), "Succeeded"},
// 		{makePodStatus("p1", v1.PodFailed, ""), "Failed"},
// 		{makePodStatus("p1", v1.PodUnknown, ""), "Unknown"},
// 		{makePodCoInitTerminated("p1"), "Init:OOMKilled"},
// 		{makePodCoInitWaiting("p1", ""), "Init:0/1"},
// 		{makePodCoInitWaiting("p2", "Waiting"), "Init:Waiting"},
// 		{makePodCoInitWaiting("p1", "PodInitializing"), "Init:0/1"},
// 		{makePodCoWaiting("p1", "Waiting"), "Waiting"},
// 		{makePodCoWaiting("p1", ""), ""},
// 		{makePodCoTerminated("p1", "OOMKilled", 0, true), Terminating},
// 		{makePodCoTerminated("p2", "OOMKilled", 0, false), "OOMKilled"},
// 		{makePodCoTerminated("p1", "", 0, true), Terminating},
// 		{makePodCoTerminated("p1", "", 0, false), "ExitCode:1"},
// 		{makePodCoTerminated("p1", "", 1, true), Terminating},
// 		{makePodCoTerminated("p1", "", 1, false), "Signal:1"},
// 	}

// 	var p Pod
// 	for _, u := range uu {
// 		assert.Equal(t, u.e, p.phase(u.p))
// 	}
// }

// func makePodStatus(n string, phase v1.PodPhase, reason string) *v1.Pod {
// 	po := makePod(n)
// 	po.Status = v1.PodStatus{
// 		Phase:  phase,
// 		Reason: reason,
// 	}

// 	return po
// }

// func makePodCoInitTerminated(n string) *v1.Pod {
// 	po := makePod(n)

// 	po.Status.InitContainerStatuses = []v1.ContainerStatus{
// 		{
// 			State: v1.ContainerState{
// 				Terminated: &v1.ContainerStateTerminated{
// 					Reason:   "OOMKilled",
// 					ExitCode: 1,
// 				},
// 			},
// 		},
// 	}

// 	return po
// }

// func makePodCoInitWaiting(n, reason string) *v1.Pod {
// 	po := makePod(n)

// 	po.Status.InitContainerStatuses = []v1.ContainerStatus{
// 		{
// 			State: v1.ContainerState{
// 				Waiting: &v1.ContainerStateWaiting{
// 					Reason: reason,
// 				},
// 			},
// 		},
// 	}

// 	return po
// }

// func makePodCoTerminated(n, reason string, signal int32, deleted bool) *v1.Pod {
// 	po := makePod(n)

// 	if deleted {
// 		po.DeletionTimestamp = &metav1.Time{Time: time.Now()}
// 	}
// 	po.Status.ContainerStatuses = []v1.ContainerStatus{
// 		{
// 			State: v1.ContainerState{
// 				Terminated: &v1.ContainerStateTerminated{
// 					Reason:   reason,
// 					Signal:   signal,
// 					ExitCode: 1,
// 				},
// 			},
// 		},
// 	}

// 	return po
// }

// func makePodCoWaiting(n, reason string) *v1.Pod {
// 	po := makePod(n)

// 	po.Status.ContainerStatuses = []v1.ContainerStatus{
// 		{
// 			State: v1.ContainerState{
// 				Waiting: &v1.ContainerStateWaiting{
// 					Reason: reason,
// 				},
// 			},
// 		},
// 	}

// 	return po
// }

// func makePod(n string) *v1.Pod {
// 	return &v1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      n,
// 			Namespace: "default",
// 		},
// 		Spec: v1.PodSpec{
// 			InitContainers: []v1.Container{
// 				{
// 					Name: "ic1",
// 				},
// 			},
// 			Containers: []v1.Container{
// 				{
// 					Name: "c1",
// 				},
// 			},
// 		},
// 	}
// }
