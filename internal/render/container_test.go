package render_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestContainer(t *testing.T) {
	var c render.Container

	cres := render.ContainerRes{
		Container: makeContainer(),
		Status:    makeContainerStatus(),
		MX:        makeContainerMetrics(),
		IsInit:    false,
		Age:       makeAge(),
	}
	var r render.Row
	assert.Nil(t, c.Render(cres, "blee", &r))
	assert.Equal(t, "fred", r.ID)
	assert.Equal(t, render.Fields{
		"fred",
		"img",
		"false",
		"Running",
		"false",
		"0",
		"off:off",
		"10",
		"20",
		"50",
		"20",
		"",
	},
		r.Fields[:len(r.Fields)-1],
	)
}

// ----------------------------------------------------------------------------
// Helpers...

func toQty(s string) resource.Quantity {
	q, _ := resource.ParseQuantity(s)
	return q

}

func makeContainerMetrics() *mv1beta1.ContainerMetrics {
	return &mv1beta1.ContainerMetrics{
		Name: "fred",
		Usage: v1.ResourceList{
			v1.ResourceCPU:    toQty("10m"),
			v1.ResourceMemory: toQty("20Mi"),
		},
	}
}

func makeAge() metav1.Time {
	return metav1.Time{Time: testTime()}
}

func makeContainer() *v1.Container {
	return &v1.Container{
		Name:  "fred",
		Image: "img",
		Resources: v1.ResourceRequirements{
			Limits: v1.ResourceList{
				v1.ResourceCPU:    toQty("20m"),
				v1.ResourceMemory: toQty("100Mi"),
			},
		},
		Env: []v1.EnvVar{
			{
				Name:  "fred",
				Value: "1",
				ValueFrom: &v1.EnvVarSource{
					ConfigMapKeyRef: &v1.ConfigMapKeySelector{Key: "blee"},
				},
			},
		},
	}
}

func makeContainerStatus() *v1.ContainerStatus {
	return &v1.ContainerStatus{
		Name:         "fred",
		State:        v1.ContainerState{Running: &v1.ContainerStateRunning{}},
		RestartCount: 0,
	}
}

func testTime() time.Time {
	t, err := time.Parse(time.RFC3339, "2018-12-14T10:36:43.326972-07:00")
	if err != nil {
		fmt.Println("TestTime Failed", err)
	}
	return t
}
