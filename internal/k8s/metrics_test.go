package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodsMetrics(t *testing.T) {
	m := NewMetricsServer(nil)

	metrics := v1beta1.PodMetricsList{
		Items: []v1beta1.PodMetrics{
			makeMxPod("p1", "1", "4Gi"),
			makeMxPod("p2", "50m", "1Mi"),
		},
	}

	mmx := make(PodsMetrics)
	m.PodsMetrics(metrics.Items, mmx)
	assert.Equal(t, 2, len(mmx))

	mx, ok := mmx["default/p1"]
	assert.True(t, ok)
	assert.Equal(t, int64(3000), mx.CurrentCPU)
	assert.Equal(t, float64(12288), mx.CurrentMEM)
}

func BenchmarkPodsMetrics(b *testing.B) {
	m := NewMetricsServer(nil)

	metrics := v1beta1.PodMetricsList{
		Items: []v1beta1.PodMetrics{
			makeMxPod("p1", "1", "4Gi"),
			makeMxPod("p2", "50m", "1Mi"),
			makeMxPod("p3", "50m", "1Mi"),
		},
	}
	mmx := make(PodsMetrics, 3)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.PodsMetrics(metrics.Items, mmx)
	}
}

func TestNodesMetrics(t *testing.T) {
	m := NewMetricsServer(nil)

	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "32", "128Gi", "50m", "2Mi"),
			makeNode("n2", "8", "4Gi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "10", "8Gi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	mmx := make(NodesMetrics)
	m.NodesMetrics(nodes.Items, metrics.Items, mmx)
	assert.Equal(t, 2, len(mmx))
	mx, ok := mmx["n1"]
	assert.True(t, ok)
	assert.Equal(t, int64(32000), mx.TotalCPU)
	assert.Equal(t, float64(131072), mx.TotalMEM)
	assert.Equal(t, int64(50), mx.AvailCPU)
	assert.Equal(t, float64(2), mx.AvailMEM)
	assert.Equal(t, int64(10000), mx.CurrentCPU)
	assert.Equal(t, float64(8192), mx.CurrentMEM)
}

func BenchmarkNodesMetrics(b *testing.B) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "100m", "4Mi", "50m", "2Mi"),
			makeNode("n2", "100m", "4Mi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "50m", "1Mi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	m := NewMetricsServer(nil)
	mmx := make(NodesMetrics)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.NodesMetrics(nodes.Items, metrics.Items, mmx)
	}
}

func TestClusterLoad(t *testing.T) {
	m := NewMetricsServer(nil)

	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "100m", "4Mi", "50m", "2Mi"),
			makeNode("n2", "100m", "4Mi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "50m", "1Mi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	mx := m.ClusterLoad(nodes.Items, metrics.Items)
	assert.Equal(t, 100.0, mx.PercCPU)
	assert.Equal(t, 50.0, mx.PercMEM)
}

func BenchmarkClusterLoad(b *testing.B) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "100m", "4Mi", "50m", "2Mi"),
			makeNode("n2", "100m", "4Mi", "50m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			makeMxNode("n1", "50m", "1Mi"),
			makeMxNode("n2", "50m", "1Mi"),
		},
	}

	m := NewMetricsServer(nil)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.ClusterLoad(nodes.Items, metrics.Items)
	}
}

func makeMxPod(name, cpu, mem string) v1beta1.PodMetrics {
	return v1beta1.PodMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: "default",
		},
		Containers: []v1beta1.ContainerMetrics{
			{Usage: makeRes(cpu, mem)},
			{Usage: makeRes(cpu, mem)},
			{Usage: makeRes(cpu, mem)},
		},
	}
}

func makeNode(name, tcpu, tmem, acpu, amem string) v1.Node {
	return v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Status: v1.NodeStatus{
			Capacity:    makeRes(tcpu, tmem),
			Allocatable: makeRes(acpu, amem),
		},
	}
}

func makeMxNode(name, cpu, mem string) v1beta1.NodeMetrics {
	return v1beta1.NodeMetrics{
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
		Usage: makeRes(cpu, mem),
	}
}

func makeRes(c, m string) v1.ResourceList {
	cpu, _ := resource.ParseQuantity(c)
	mem, _ := resource.ParseQuantity(m)

	return v1.ResourceList{
		v1.ResourceCPU:    cpu,
		v1.ResourceMemory: mem,
	}
}
