package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	res "k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodRender(t *testing.T) {
	pom := podMetrics{load(t, "po"), makePodMX("nginx", "10m", "10Mi")}

	var po render.Pod
	r := render.NewRow(12)
	err := po.Render(pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "1/1", "Running", "0", "10", "10", "10", "14", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:12])
}

func TestPodInitRender(t *testing.T) {
	pom := podMetrics{load(t, "po_init"), makePodMX("nginx", "10m", "10Mi")}

	var po render.Pod
	r := render.NewRow(12)
	err := po.Render(pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "default/nginx", r.ID)
	e := render.Fields{"default", "nginx", "1/1", "Init:0/1", "0", "10", "10", "10", "14", "172.17.0.6", "minikube", "BE"}
	assert.Equal(t, e, r.Fields[:12])
}

// ----------------------------------------------------------------------------
// Helpers...

type podMetrics struct {
	o *unstructured.Unstructured
	m *mv1beta1.PodMetrics
}

func (p podMetrics) Object() runtime.Object {
	return p.o
}

func (p podMetrics) Metrics() *mv1beta1.PodMetrics {
	return p.m
}

func makePodMX(name, cpu, mem string) *mv1beta1.PodMetrics {
	return &mv1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []mv1beta1.ContainerMetrics{
			{Usage: makeRes(cpu, mem)},
		},
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := res.ParseQuantity(c)
	mem, _ := res.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}
