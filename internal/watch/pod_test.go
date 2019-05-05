package watch

import (
	"fmt"
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func BenchmarkPodFields(b *testing.B) {
	p := NewPod(nil, "")
	po := makePod()
	ff := make(Row, podCols)

	b.ResetTimer()
	b.ReportAllocs()

	for n := 0; n < b.N; n++ {
		p.fields("", po, ff)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}

func makePod() *v1.Pod {
	var i int32 = 1
	var t = v1.HostPathDirectory
	return &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Namespace:         "blee",
			Name:              "fred",
			Labels:            map[string]string{"blee": "duh"},
			CreationTimestamp: metav1.Time{Time: testTime()},
		},
		Spec: v1.PodSpec{
			Priority:          &i,
			PriorityClassName: "bozo",
			Containers: []v1.Container{
				{
					Name:  "fred",
					Image: "blee",
					Env: []v1.EnvVar{
						{
							Name:  "fred",
							Value: "1",
							ValueFrom: &v1.EnvVarSource{
								ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
							},
						},
					},
				},
			},
			Volumes: []v1.Volume{
				{
					Name: "fred",
					VolumeSource: v1.VolumeSource{
						HostPath: &v1.HostPathVolumeSource{
							Path: "/blee",
							Type: &t,
						},
					},
				},
			},
		},
		Status: v1.PodStatus{
			Phase: "Running",
			ContainerStatuses: []v1.ContainerStatus{
				{
					Name:         "fred",
					State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
					RestartCount: 0,
				},
			},
		},
	}
}
