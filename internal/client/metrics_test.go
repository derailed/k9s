package client_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	v1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

func TestPodsMetrics(t *testing.T) {
	uu := map[string]struct {
		metrics *mv1beta1.PodMetricsList
		eSize   int
		e       client.PodsMetrics
	}{
		"dud": {
			eSize: 0,
		},

		"ok": {
			metrics: &v1beta1.PodMetricsList{
				Items: []v1beta1.PodMetrics{
					*makeMxPod("p1", "1", "4Gi"),
					*makeMxPod("p2", "50m", "1Mi"),
				},
			},
			eSize: 2,
			e: client.PodsMetrics{
				"default/p1": client.PodMetrics{
					CurrentCPU: int64(3000),
					CurrentMEM: float64(12288),
				},
			},
		},
	}

	m := client.NewMetricsServer(nil)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			mmx := make(client.PodsMetrics)
			m.PodsMetrics(u.metrics, mmx)

			assert.Equal(t, u.eSize, len(mmx))
			if u.eSize == 0 {
				return
			}
			mx, ok := mmx["default/p1"]
			assert.True(t, ok)
			assert.Equal(t, u.e["default/p1"], mx)
		})
	}
}

func BenchmarkPodsMetrics(b *testing.B) {
	m := client.NewMetricsServer(nil)

	metrics := v1beta1.PodMetricsList{
		Items: []v1beta1.PodMetrics{
			*makeMxPod("p1", "1", "4Gi"),
			*makeMxPod("p2", "50m", "1Mi"),
			*makeMxPod("p3", "50m", "1Mi"),
		},
	}
	mmx := make(client.PodsMetrics, 3)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.PodsMetrics(&metrics, mmx)
	}
}

func TestNodesMetrics(t *testing.T) {
	uu := map[string]struct {
		nodes   *v1.NodeList
		metrics *mv1beta1.NodeMetricsList
		eSize   int
		e       client.NodesMetrics
	}{
		"duds": {
			eSize: 0,
		},
		"no_nodes": {
			metrics: &v1beta1.NodeMetricsList{
				Items: []v1beta1.NodeMetrics{
					*makeMxNode("n1", "10", "8Gi"),
					*makeMxNode("n2", "50m", "1Mi"),
				},
			},
			eSize: 0,
		},
		"no_metrics": {
			nodes: &v1.NodeList{
				Items: []v1.Node{
					makeNode("n1", "32", "128Gi", "50m", "2Mi"),
					makeNode("n2", "8", "4Gi", "50m", "10Mi"),
				},
			},
			eSize: 0,
		},
		"ok": {
			nodes: &v1.NodeList{
				Items: []v1.Node{
					makeNode("n1", "32", "128Gi", "50m", "2Mi"),
					makeNode("n2", "8", "4Gi", "50m", "10Mi"),
				},
			},
			metrics: &v1beta1.NodeMetricsList{
				Items: []v1beta1.NodeMetrics{
					*makeMxNode("n1", "10", "8Gi"),
					*makeMxNode("n2", "50m", "1Mi"),
				},
			},
			eSize: 2,
			e: client.NodesMetrics{
				"n1": client.NodeMetrics{
					TotalCPU: int64(32000),
					TotalMEM: float64(131072),
					AvailCPU: int64(50),
					AvailMEM: float64(2),
					CurrentMetrics: client.CurrentMetrics{
						CurrentCPU: int64(10000),
						CurrentMEM: float64(8192),
					},
				},
			},
		},
	}

	m := client.NewMetricsServer(nil)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			mmx := make(client.NodesMetrics)
			m.NodesMetrics(u.nodes, u.metrics, mmx)

			assert.Equal(t, u.eSize, len(mmx))
			if u.eSize == 0 {
				return
			}
			mx, ok := mmx["n1"]
			assert.True(t, ok)
			assert.Equal(t, u.e["n1"], mx)
		})
	}
}

func BenchmarkNodesMetrics(b *testing.B) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "100m", "4Mi", "100m", "2Mi"),
			makeNode("n2", "100m", "4Mi", "100m", "2Mi"),
		},
	}

	metrics := v1beta1.NodeMetricsList{
		Items: []v1beta1.NodeMetrics{
			*makeMxNode("n1", "50m", "1Mi"),
			*makeMxNode("n2", "50m", "1Mi"),
		},
	}

	m := client.NewMetricsServer(nil)
	mmx := make(client.NodesMetrics)

	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.NodesMetrics(&nodes, &metrics, mmx)
	}
}

func TestClusterLoad(t *testing.T) {
	uu := map[string]struct {
		nodes   *v1.NodeList
		metrics *mv1beta1.NodeMetricsList
		eSize   int
		e       client.ClusterMetrics
	}{
		"duds": {
			eSize: 0,
		},
		"no_nodes": {
			metrics: &v1beta1.NodeMetricsList{
				Items: []v1beta1.NodeMetrics{
					*makeMxNode("n1", "10", "8Gi"),
					*makeMxNode("n2", "50m", "1Mi"),
				},
			},
			eSize: 0,
		},
		"no_metrics": {
			nodes: &v1.NodeList{
				Items: []v1.Node{
					makeNode("n1", "32", "128Gi", "50m", "2Mi"),
					makeNode("n2", "8", "4Gi", "50m", "10Mi"),
				},
			},
			eSize: 0,
		},
		"ok": {

			nodes: &v1.NodeList{
				Items: []v1.Node{
					makeNode("n1", "100m", "4Mi", "50m", "2Mi"),
					makeNode("n2", "100m", "4Mi", "50m", "2Mi"),
				},
			},
			metrics: &v1beta1.NodeMetricsList{
				Items: []v1beta1.NodeMetrics{
					*makeMxNode("n1", "50m", "1Mi"),
					*makeMxNode("n2", "50m", "1Mi"),
				},
			},
			eSize: 2,
			e: client.ClusterMetrics{
				PercCPU: 100.0,
				PercMEM: 50.0,
			},
		},
	}

	m := client.NewMetricsServer(nil)
	for k := range uu {
		u := uu[k]
		t.Run(k, func(t *testing.T) {
			var cmx client.ClusterMetrics
			m.ClusterLoad(u.nodes, u.metrics, &cmx)

			assert.Equal(t, u.e, cmx)
		})
	}
}

func BenchmarkClusterLoad(b *testing.B) {
	nodes := v1.NodeList{
		Items: []v1.Node{
			makeNode("n1", "100m", "4Mi", "50m", "2Mi"),
			makeNode("n2", "100m", "4Mi", "50m", "2Mi"),
		},
	}

	metrics := mv1beta1.NodeMetricsList{
		Items: []mv1beta1.NodeMetrics{
			*makeMxNode("n1", "50m", "1Mi"),
			*makeMxNode("n2", "50m", "1Mi"),
		},
	}

	m := client.NewMetricsServer(nil)
	var mx client.ClusterMetrics
	b.ResetTimer()
	b.ReportAllocs()
	for n := 0; n < b.N; n++ {
		m.ClusterLoad(&nodes, &metrics, &mx)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func makeMxPod(name, cpu, mem string) *v1beta1.PodMetrics {
	return &v1beta1.PodMetrics{
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

func makeMxNode(name, cpu, mem string) *v1beta1.NodeMetrics {
	return &v1beta1.NodeMetrics{
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
