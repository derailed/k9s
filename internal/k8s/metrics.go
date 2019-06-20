package k8s

import (
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// MetricsServer serves cluster metrics for nodes and pods.
	MetricsServer struct {
		*base
		Connection
	}

	currentMetrics struct {
		CurrentCPU int64
		CurrentMEM float64
	}

	// PodMetrics represent an aggregation of all pod containers metrics.
	PodMetrics currentMetrics

	// NodeMetrics describes raw node metrics.
	NodeMetrics struct {
		currentMetrics
		AvailCPU int64
		AvailMEM float64
		TotalCPU int64
		TotalMEM float64
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
	return &MetricsServer{&base{}, c}
}

// NodesMetrics retrieves metrics for a given set of nodes.
func (m *MetricsServer) NodesMetrics(nodes Collection, metrics *mv1beta1.NodeMetricsList, mmx NodesMetrics) {
	for _, n := range nodes {
		no := n.(*v1.Node)
		mmx[no.Name] = NodeMetrics{
			AvailCPU: no.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: ToMB(no.Status.Allocatable.Memory().Value()),
			TotalCPU: no.Status.Capacity.Cpu().MilliValue(),
			TotalMEM: ToMB(no.Status.Capacity.Memory().Value()),
		}
	}

	for _, c := range metrics.Items {
		if mx, ok := mmx[c.Name]; ok {
			mx.CurrentCPU = c.Usage.Cpu().MilliValue()
			mx.CurrentMEM = ToMB(c.Usage.Memory().Value())
			mmx[c.Name] = mx
		}
	}
}

// ClusterLoad retrieves all cluster nodes metrics.
func (m *MetricsServer) ClusterLoad(nos Collection, nmx Collection, mx *ClusterMetrics) {
	nodeMetrics := make(NodesMetrics, len(nos))
	for _, n := range nos {
		no := n.(*v1.Node)
		nodeMetrics[no.Name] = NodeMetrics{
			AvailCPU: no.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: ToMB(no.Status.Allocatable.Memory().Value()),
		}
	}

	for _, mx := range nmx {
		mxx := mx.(*mv1beta1.NodeMetrics)
		if m, ok := nodeMetrics[mxx.Name]; ok {
			m.CurrentCPU = mxx.Usage.Cpu().MilliValue()
			m.CurrentMEM = ToMB(mxx.Usage.Memory().Value())
			nodeMetrics[mxx.Name] = m
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
}

// FetchNodesMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchNodesMetrics() (*mv1beta1.NodeMetricsList, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	return client.Metrics().NodeMetricses().List(metav1.ListOptions{})
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodsMetrics(ns string) (*mv1beta1.PodMetricsList, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}

	return client.Metrics().PodMetricses(ns).List(metav1.ListOptions{})
}

// PodsMetrics retrieves metrics for all pods in a given namespace.
func (m *MetricsServer) PodsMetrics(pods *mv1beta1.PodMetricsList, mmx PodsMetrics) {
	// Compute all pod's containers metrics.
	for _, p := range pods.Items {
		var mx PodMetrics
		for _, c := range p.Containers {
			mx.CurrentCPU += c.Usage.Cpu().MilliValue()
			mx.CurrentMEM += ToMB(c.Usage.Memory().Value())
		}
		mmx[p.Namespace+"/"+p.Name] = mx
	}
}
