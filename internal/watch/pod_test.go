package watch

import (
	"testing"

	"gotest.tools/assert"
)

func TestPodList(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPod(cmo, "")

	o := no.List("")
	assert.Assert(t, o == nil)
}

func TestPodGet(t *testing.T) {
	cmo := NewMockConnection()
	no := NewPod(cmo, "")

	o, err := no.Get("")
	assert.ErrorContains(t, err, "not found")
	assert.Assert(t, o == nil)
}

// ----------------------------------------------------------------------------
// Helpers...

// func makePod() *v1.Pod {
// 	var i int32 = 1
// 	var t = v1.HostPathDirectory
// 	return &v1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Namespace:         "blee",
// 			Name:              "fred",
// 			Labels:            map[string]string{"blee": "duh"},
// 			CreationTimestamp: metav1.Time{Time: testTime()},
// 		},
// 		Spec: v1.PodSpec{
// 			Priority:          &i,
// 			PriorityClassName: "bozo",
// 			Containers: []v1.Container{
// 				{
// 					Name:  "fred",
// 					Image: "blee",
// 					Env: []v1.EnvVar{
// 						{
// 							Name:  "fred",
// 							Value: "1",
// 							ValueFrom: &v1.EnvVarSource{
// 								ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
// 							},
// 						},
// 					},
// 				},
// 			},
// 			Volumes: []v1.Volume{
// 				{
// 					Name: "fred",
// 					VolumeSource: v1.VolumeSource{
// 						HostPath: &v1.HostPathVolumeSource{
// 							Path: "/blee",
// 							Type: &t,
// 						},
// 					},
// 				},
// 			},
// 		},
// 		Status: v1.PodStatus{
// 			Phase: "Running",
// 			ContainerStatuses: []v1.ContainerStatus{
// 				{
// 					Name:         "fred",
// 					State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
// 					RestartCount: 0,
// 				},
// 			},
// 		},
// 	}
// }
