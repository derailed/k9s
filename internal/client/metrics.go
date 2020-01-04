package client

import (
	"math"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

// MetricsServer serves cluster metrics for nodes and pods.
type MetricsServer struct {
	Connection
}

// NewMetricsServer return a metric server instance.
func NewMetricsServer(c Connection) *MetricsServer {
	return &MetricsServer{Connection: c}
}

// NodesMetrics retrieves metrics for a given set of nodes.
func (m *MetricsServer) NodesMetrics(nodes *v1.NodeList, metrics *mv1beta1.NodeMetricsList, mmx NodesMetrics) {
	for _, no := range nodes.Items {
		mmx[no.Name] = NodeMetrics{
			AvailCPU: no.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: toMB(no.Status.Allocatable.Memory().Value()),
			TotalCPU: no.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: toMB(no.Status.Capacity.Memory().Value()),
		}
	}

	for _, c := range metrics.Items {
		if mx, ok := mmx[c.Name]; ok {
			mx.CurrentCPU = c.Usage.Cpu().MilliValue()
			mx.CurrentMEM = toMB(c.Usage.Memory().Value())
			mmx[c.Name] = mx
		}
	}
}

// ClusterLoad retrieves all cluster nodes metrics.
func (m *MetricsServer) ClusterLoad(nos *v1.NodeList, nmx *mv1beta1.NodeMetricsList, mx *ClusterMetrics) error {
	nodeMetrics := make(NodesMetrics, len(nos.Items))
	for _, no := range nos.Items {
		nodeMetrics[no.Name] = NodeMetrics{
			AvailCPU: no.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: toMB(no.Status.Allocatable.Memory().Value()),
		}
	}

	for _, mx := range nmx.Items {
		if m, ok := nodeMetrics[mx.Name]; ok {
			m.CurrentCPU = mx.Usage.Cpu().MilliValue()
			m.CurrentMEM = toMB(mx.Usage.Memory().Value())
			nodeMetrics[mx.Name] = m
		}
	}

	var cpu, tcpu, mem, tmem float64
	for _, mx := range nodeMetrics {
		cpu += float64(mx.CurrentCPU)
		tcpu += float64(mx.AvailCPU)
		mem += mx.CurrentMEM
		tmem += mx.AvailMEM
	}
	mx.PercCPU, mx.PercMEM = toPerc(cpu, tcpu), toPerc(mem, tmem)

	return nil
}

// FetchNodesMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchNodesMetrics() (*mv1beta1.NodeMetricsList, error) {
	auth, err := m.CanI("", "metrics.k8s.io/v1beta1/nodes", []string{"list"})
	if !auth || err != nil {
		return nil, err
	}

	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	return client.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodsMetrics(ns string) (*mv1beta1.PodMetricsList, error) {
	if ns == NamespaceAll {
		ns = AllNamespaces
	}

	auth, err := m.CanI(ns, "metrics.k8s.io/v1beta1/pods", []string{"list"})
	if !auth || err != nil {
		return &mv1beta1.PodMetricsList{}, err
	}

	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	return client.MetricsV1beta1().PodMetricses(ns).List(metav1.ListOptions{})
}

// FetchPodMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodMetrics(ns, sel string) (*mv1beta1.PodMetrics, error) {
	if ns == NamespaceAll {
		ns = AllNamespaces
	}
	auth, err := m.CanI(ns, "metrics.k8s.io/v1beta1/pods", []string{"get"})
	if !auth || err != nil {
		return nil, err
	}

	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	return client.MetricsV1beta1().PodMetricses(ns).Get(sel, metav1.GetOptions{})
}

// PodsMetrics retrieves metrics for all pods in a given namespace.
func (m *MetricsServer) PodsMetrics(pods *mv1beta1.PodMetricsList, mmx PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods.Items {
		var mx PodMetrics
		for _, c := range p.Containers {
			mx.CurrentCPU += c.Usage.Cpu().MilliValue()
			mx.CurrentMEM += toMB(c.Usage.Memory().Value())
		}
		mmx[p.Namespace+"/"+p.Name] = mx
	}
}

// 0---------------------------------------------------------------------------
// Helpers...

const megaByte = 1024 * 1024

// toMB converts bytes to megabytes.
func toMB(v int64) float64 {
	return float64(v) / megaByte
}

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}
