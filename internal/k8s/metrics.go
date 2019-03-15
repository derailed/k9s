package k8s

import (
	"fmt"
	"math"
	"path"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	metricsV1beta1api "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

type (
	// MetricsServer serves cluster metrics for nodes and pods.
	MetricsServer struct{}

	// Metric tracks resource metrics.
	Metric struct {
		CPU      string
		Mem      string
		AvailCPU string
		AvailMem string
	}
)

// NewMetricsServer return a metric server instance.
func NewMetricsServer() *MetricsServer {
	return &MetricsServer{}
}

// NodeMetrics retrieves all nodes metrics
func (m *MetricsServer) NodeMetrics() (Metric, error) {
	var mx Metric

	opts := metav1.ListOptions{}
	nn, err := conn.dialOrDie().CoreV1().Nodes().List(opts)
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	nods := make([]string, len(nn.Items))
	var maxCPU, maxMem float64
	for i, n := range nn.Items {
		nods[i] = n.Name
		c := n.Status.Allocatable["cpu"]
		maxCPU += float64(c.MilliValue())
		m := n.Status.Allocatable["memory"]
		maxMem += float64(m.Value() / (1024 * 1024))
	}

	mm, err := m.getNodeMetrics()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	var cpu, mem float64
	for _, n := range nods {
		for _, m := range mm.Items {
			if m.Name == n {
				cpu += float64(m.Usage.Cpu().MilliValue())
				mem += float64(m.Usage.Memory().Value() / (1024 * 1024))
			}
		}
	}
	mx = Metric{
		CPU: fmt.Sprintf("%0.f%%", math.Round((cpu/maxCPU)*100)),
		Mem: fmt.Sprintf("%0.f%%", math.Round((mem/maxMem)*100)),
	}

	return mx, nil
}

// PodMetrics retrieves all pods metrics
func (m *MetricsServer) PodMetrics() (map[string]Metric, error) {
	mx := map[string]Metric{}

	mm, err := m.getPodMetrics()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	for _, m := range mm.Items {
		var cpu, mem int64
		for _, c := range m.Containers {
			cpu += c.Usage.Cpu().MilliValue()
			mem += c.Usage.Memory().Value() / (1024 * 1024)
		}
		pa := path.Join(m.Namespace, m.Name)
		mx[pa] = Metric{CPU: fmt.Sprintf("%dm", cpu), Mem: fmt.Sprintf("%dMi", mem)}
	}

	return mx, nil
}

// PerNodeMetrics retrieves all nodes metrics
func (m *MetricsServer) PerNodeMetrics(nn []v1.Node) (map[string]Metric, error) {
	mx := map[string]Metric{}

	mm, err := m.getNodeMetrics()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	for _, n := range nn {
		acpu := n.Status.Allocatable["cpu"]
		amem := n.Status.Allocatable["memory"]
		var cpu, mem int64
		for _, m := range mm.Items {
			if m.Name == n.Name {
				cpu += m.Usage.Cpu().MilliValue()
				mem += m.Usage.Memory().Value() / (1024 * 1024)
			}
		}
		mx[n.Name] = Metric{
			CPU:      fmt.Sprintf("%dm", cpu),
			Mem:      fmt.Sprintf("%dMi", mem),
			AvailCPU: fmt.Sprintf("%dm", acpu.MilliValue()),
			AvailMem: fmt.Sprintf("%dMi", amem.Value()/(1024*1024)),
		}
	}

	return mx, nil
}

func (m *MetricsServer) getPodMetrics() (*metricsapi.PodMetricsList, error) {
	if conn.hasMetricsServer() {
		return m.podMetricsViaService()
	}

	var mx *metricsapi.PodMetricsList
	conn, err := conn.heapsterDial()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	return conn.GetPodMetrics("", "", true, labels.Everything())
}

func (m *MetricsServer) getNodeMetrics() (*metricsapi.NodeMetricsList, error) {
	if conn.hasMetricsServer() {
		return m.nodeMetricsViaService()
	}

	var mx *metricsapi.NodeMetricsList
	conn, err := conn.heapsterDial()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	return conn.GetNodeMetrics("", labels.Everything().String())
}

func (*MetricsServer) nodeMetricsViaService() (*metricsapi.NodeMetricsList, error) {
	var mx *metricsapi.NodeMetricsList

	clt, err := conn.mxsDial()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	selector := labels.Everything()
	var versionedMetrics *metricsV1beta1api.NodeMetricsList
	mc := clt.Metrics()
	nm := mc.NodeMetricses()
	versionedMetrics, err = nm.List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Warn().Msgf("%s", err)
		return nil, err
	}

	metrics := &metricsapi.NodeMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_NodeMetricsList_To_metrics_NodeMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		log.Warn().Msgf("%s", err)
		return nil, err
	}

	return metrics, nil
}

func (*MetricsServer) podMetricsViaService() (*metricsapi.PodMetricsList, error) {
	var mx *metricsapi.PodMetricsList

	clt, err := conn.mxsDial()
	if err != nil {
		log.Warn().Msgf("%s", err)
		return mx, err
	}

	selector := labels.Everything()
	var versionedMetrics *metricsV1beta1api.PodMetricsList
	mc := clt.Metrics()
	nm := mc.PodMetricses("")
	versionedMetrics, err = nm.List(metav1.ListOptions{LabelSelector: selector.String()})
	if err != nil {
		log.Warn().Msgf("%s", err)
		return nil, err
	}

	metrics := &metricsapi.PodMetricsList{}
	err = metricsV1beta1api.Convert_v1beta1_PodMetricsList_To_metrics_PodMetricsList(versionedMetrics, metrics, nil)
	if err != nil {
		log.Warn().Msgf("%s", err)
		return nil, err
	}

	return metrics, nil
}
