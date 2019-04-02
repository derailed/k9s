package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// MetricsServer serves cluster metrics for nodes and pods.
	MetricsServer struct {
		Connection
	}

	// NodeMetrics describes raw node metrics.
	NodeMetrics struct {
		CurrentCPU int64
		CurrentMEM float64
		AvailCPU   int64
		AvailMEM   float64
		TotalCPU   int64
		TotalMEM   float64
	}

	// PodMetrics represent an aggregation of all pod containers metrics.
	PodMetrics struct {
		CurrentCPU int64
		CurrentMEM float64
	}

	// ClusterMetrics summarizes total node metrics as percentages.
	ClusterMetrics struct {
		PercCPU float64
		PercMEM float64
	}

	// NodesMetrics tracks usage metrics per nodes.
	NodesMetrics map[string]NodeMetrics

	// PodsMetrics tracks usage metrics per pods.
	PodsMetrics map[string]PodMetrics
)

// NewMetricsServer return a metric server instance.
func NewMetricsServer(c Connection) *MetricsServer {
	return &MetricsServer{c}
}

// NodesMetrics retrieves metrics for a given set of nodes.
func (m *MetricsServer) NodesMetrics(nodes []v1.Node, metrics []mv1beta1.NodeMetrics, mmx NodesMetrics) {
	for _, n := range nodes {
		mmx[n.Name] = NodeMetrics{
			AvailCPU: n.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: asMi(n.Status.Allocatable.Memory().Value()),
			TotalCPU: n.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: asMi(n.Status.Capacity.Memory().Value()),
		}
	}

	for _, c := range metrics {
		if mx, ok := mmx[c.Name]; ok {
			mx.CurrentCPU = c.Usage.Cpu().MilliValue()
			mx.CurrentMEM = asMi(c.Usage.Memory().Value())
			mmx[c.Name] = mx
		}
	}
}

// ClusterLoad retrieves all cluster nodes metrics.
func (m *MetricsServer) ClusterLoad(nodes []v1.Node, metrics []mv1beta1.NodeMetrics) ClusterMetrics {
	nodeMetrics := make(NodesMetrics, len(nodes))

	for _, n := range nodes {
		nodeMetrics[n.Name] = NodeMetrics{
			AvailCPU: n.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: asMi(n.Status.Allocatable.Memory().Value()),
			TotalCPU: n.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: asMi(n.Status.Capacity.Memory().Value()),
		}
	}

	for _, mx := range metrics {
		if m, ok := nodeMetrics[mx.Name]; ok {
			m.CurrentCPU = mx.Usage.Cpu().MilliValue()
			m.CurrentMEM = asMi(mx.Usage.Memory().Value())
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

	return ClusterMetrics{PercCPU: toPerc(cpu, tcpu), PercMEM: toPerc(mem, tmem)}
}

// FetchNodesMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchNodesMetrics() ([]mv1beta1.NodeMetrics, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	list, err := client.Metrics().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodsMetrics(ns string) ([]mv1beta1.PodMetrics, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	list, err := client.Metrics().PodMetricses(ns).List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return list.Items, nil
}

// PodsMetrics retrieves metrics for all pods in a given namespace.
func (m *MetricsServer) PodsMetrics(pods []mv1beta1.PodMetrics, mmx PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods {
		var mx PodMetrics
		for _, c := range p.Containers {
			mx.CurrentCPU += c.Usage.Cpu().MilliValue()
			mx.CurrentMEM += asMi(c.Usage.Memory().Value())
		}
		mmx[p.Namespace+"/"+p.Name] = mx
	}
}
