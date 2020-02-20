package client

import (
	"fmt"
	"math"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	mxCacheSize   = 100
	mxCacheExpiry = 1 * time.Minute
)

// MetricsDial tracks global metric server handle.
var MetricsDial *MetricsServer

// DialMetrics dials the metrics server.
func DialMetrics(c Connection) *MetricsServer {
	if MetricsDial == nil {
		MetricsDial = NewMetricsServer(c)
	}

	return MetricsDial
}

// ResetMetrics resets the metric server handle.
func ResetMetrics() {
	MetricsDial = nil
}

// MetricsServer serves cluster metrics for nodes and pods.
type MetricsServer struct {
	Connection

	cache *cache.LRUExpireCache
}

// NewMetricsServer return a metric server instance.
func NewMetricsServer(c Connection) *MetricsServer {
	return &MetricsServer{
		Connection: c,
		cache:      cache.NewLRUExpireCache(mxCacheSize),
	}
}

// NodesMetrics retrieves metrics for a given set of nodes.
func (m *MetricsServer) NodesMetrics(nodes *v1.NodeList, metrics *mv1beta1.NodeMetricsList, mmx NodesMetrics) {
	if nodes == nil || metrics == nil {
		return
	}

	for _, no := range nodes.Items {
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
func (m *MetricsServer) ClusterLoad(nos *v1.NodeList, nmx *mv1beta1.NodeMetricsList, mx *ClusterMetrics) error {
	if nos == nil || nmx == nil {
		return fmt.Errorf("invalid node or node metrics lists")
	}
	nodeMetrics := make(NodesMetrics, len(nos.Items))
	for _, no := range nos.Items {
		nodeMetrics[no.Name] = NodeMetrics{
			AvailCPU: no.Status.Allocatable.Cpu().MilliValue(),
			AvailMEM: ToMB(no.Status.Allocatable.Memory().Value()),
		}
	}
	for _, mx := range nmx.Items {
		if m, ok := nodeMetrics[mx.Name]; ok {
			m.CurrentCPU = mx.Usage.Cpu().MilliValue()
			m.CurrentMEM = ToMB(mx.Usage.Memory().Value())
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

func (m *MetricsServer) checkAccess(ns, gvr, msg string) error {
	if !m.HasMetrics() {
		return fmt.Errorf("No metrics-server detected on cluster")
	}

	auth, err := m.CanI(ns, gvr, ListAccess)
	if err != nil {
		return err
	}
	if !auth {
		return fmt.Errorf(msg)
	}
	return nil
}

// FetchNodesMetrics return all metrics for nodes.
func (m *MetricsServer) FetchNodesMetrics() (*mv1beta1.NodeMetricsList, error) {
	const msg = "user is not authorized to list node metrics"

	mx := new(mv1beta1.NodeMetricsList)
	if err := m.checkAccess("", "metrics.k8s.io/v1beta1/nodes", msg); err != nil {
		return mx, err
	}

	const key = "nodes"
	if entry, ok := m.cache.Get(key); ok && entry != nil {
		mxList, ok := entry.(*mv1beta1.NodeMetricsList)
		if !ok {
			return nil, fmt.Errorf("expected nodemetricslist but got %T", entry)
		}
		return mxList, nil
	}

	client, err := m.MXDial()
	if err != nil {
		return mx, err
	}
	mxList, err := client.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
	if err != nil {
		return mx, err
	}
	m.cache.Add(key, mxList, mxCacheExpiry)

	return mxList, nil
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodsMetrics(ns string) (*mv1beta1.PodMetricsList, error) {
	mx := new(mv1beta1.PodMetricsList)
	const msg = "user is not authorized to list pods metrics"

	if ns == NamespaceAll {
		ns = AllNamespaces
	}
	if err := m.checkAccess(ns, "metrics.k8s.io/v1beta1/pods", msg); err != nil {
		return mx, err
	}

	key := FQN(ns, "pods")
	if entry, ok := m.cache.Get(key); ok {
		mxList, ok := entry.(*mv1beta1.PodMetricsList)
		if !ok {
			return mx, fmt.Errorf("expected podmetricslist but got %T", entry)
		}
		return mxList, nil
	}

	client, err := m.MXDial()
	if err != nil {
		return mx, err
	}
	mxList, err := client.MetricsV1beta1().PodMetricses(ns).List(metav1.ListOptions{})
	if err != nil {
		return mx, err
	}
	m.cache.Add(key, mxList, mxCacheExpiry)

	return mxList, err
}

// FetchPodMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodMetrics(fqn string) (*mv1beta1.PodMetrics, error) {
	var mx *mv1beta1.PodMetrics
	const msg = "user is not authorized to list pod metrics"

	ns, n := Namespaced(fqn)
	if ns == NamespaceAll {
		ns = AllNamespaces
	}
	if err := m.checkAccess(ns, "metrics.k8s.io/v1beta1/pods", msg); err != nil {
		return mx, err
	}

	var key = FQN(ns, "pods")
	if entry, ok := m.cache.Get(key); ok {
		if list, ok := entry.(*mv1beta1.PodMetricsList); ok && list != nil {
			for _, m := range list.Items {
				if FQN(m.Namespace, m.Name) == fqn {
					return &m, nil
				}
			}
		}
	}

	client, err := m.MXDial()
	if err != nil {
		return mx, err
	}
	mx, err = client.MetricsV1beta1().PodMetricses(ns).Get(n, metav1.GetOptions{})
	if err != nil {
		return mx, err
	}
	m.cache.Add(key, mx, mxCacheExpiry)

	return mx, nil
}

// PodsMetrics retrieves metrics for all pods in a given namespace.
func (m *MetricsServer) PodsMetrics(pods *mv1beta1.PodMetricsList, mmx PodsMetrics) {
	if pods == nil {
		return
	}

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

// 0---------------------------------------------------------------------------
// Helpers...

const megaByte = 1024 * 1024

// ToMB converts bytes to megabytes.
func ToMB(v int64) float64 {
	return float64(v) / megaByte
}

func toPerc(v1, v2 float64) float64 {
	if v2 == 0 {
		return 0
	}
	return math.Round((v1 / v2) * 100)
}
