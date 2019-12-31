package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestNodeRender(t *testing.T) {
	pom := nodeMetrics{
		load(t, "no"),
		makeNodeMX("n1", "10m", "10Mi"),
		[]*v1.Pod{},
	}

	var no render.Node
	r := render.NewRow(14)
	err := no.Render(pom, "", &r)
	assert.Nil(t, err)

	assert.Equal(t, "minikube", r.ID)
	e := render.Fields{"minikube", "Ready", "master", "v1.15.2", "4.15.0", "192.168.64.107", "<none>", "10", "10", "0", "0", "4000", "7874"}
	assert.Equal(t, e, r.Fields[:13])
}

// ----------------------------------------------------------------------------
// Helpers...

type nodeMetrics struct {
	o   *unstructured.Unstructured
	m   *mv1beta1.NodeMetrics
	pod []*v1.Pod
}

func (p nodeMetrics) Object() runtime.Object {
	return p.o
}

func (p nodeMetrics) Metrics() *mv1beta1.NodeMetrics {
	return p.m
}

func (p nodeMetrics) Pods() []*v1.Pod {
	return p.pod
}

func makeNodeMX(name, cpu, mem string) *mv1beta1.NodeMetrics {
	return &mv1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Usage: makeRes(cpu, mem),
	}
}
