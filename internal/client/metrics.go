// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	mxCacheSize           = 100
	mxCacheExpiry         = 1 * time.Minute
	metricsPageSize int64 = 500

	// mxFetchWait caps how long callers block on a cold metrics fetch before
	// falling back to the last known metrics while the fetch completes in the
	// background. Listing pod metrics across all namespaces on large clusters
	// can take tens of seconds and must not stall the views refresh loop.
	mxFetchWait = 500 * time.Millisecond
)

// ErrMetricsNotReady indicates metrics have not been fetched yet for a given scope.
var ErrMetricsNotReady = errors.New("metrics not yet available")

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

	cache    *cache.LRUExpireCache
	fmx      sync.Mutex
	inflight map[string]chan struct{}
	stale    map[string]any
}

// NewMetricsServer return a metric server instance.
func NewMetricsServer(c Connection) *MetricsServer {
	return &MetricsServer{
		Connection: c,
		cache:      cache.NewLRUExpireCache(mxCacheSize),
		inflight:   make(map[string]chan struct{}),
		stale:      make(map[string]any),
	}
}

// fetchList returns cached metrics for a given key when fresh. On a cache miss,
// the list is refreshed in the background (a single fetch per key at a time)
// and the last known metrics are returned after a short grace period so callers
// never stall on slow metrics api calls.
func (m *MetricsServer) fetchList(ctx context.Context, key string, list func(context.Context) (any, error)) (any, error) {
	if v, ok := m.cache.Get(key); ok {
		return v, nil
	}

	m.fmx.Lock()
	done, ok := m.inflight[key]
	if !ok {
		done = make(chan struct{})
		m.inflight[key] = done
		timeout := DefaultCallTimeoutDuration
		if m.Connection != nil {
			if cfg := m.Config(); cfg != nil {
				timeout = cfg.CallTimeout()
			}
		}
		go func() {
			defer close(done)
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			v, err := list(ctx)
			m.fmx.Lock()
			defer m.fmx.Unlock()
			delete(m.inflight, key)
			if err != nil {
				slog.Warn("Metrics fetch failed",
					slogs.Key, key,
					slogs.Error, err,
				)
				return
			}
			m.cache.Add(key, v, mxCacheExpiry)
			m.stale[key] = v
		}()
	}
	m.fmx.Unlock()

	select {
	case <-done:
	case <-ctx.Done():
	case <-time.After(mxFetchWait):
	}

	if v, ok := m.cache.Get(key); ok {
		return v, nil
	}
	m.fmx.Lock()
	v, ok := m.stale[key]
	m.fmx.Unlock()
	if ok {
		return v, nil
	}

	return nil, ErrMetricsNotReady
}

// ClusterLoad retrieves all cluster nodes metrics.
func (*MetricsServer) ClusterLoad(nos *v1.NodeList, nmx *mv1beta1.NodeMetricsList, mx *ClusterMetrics) error {
	if nos == nil || nmx == nil {
		return fmt.Errorf("invalid node or node metrics lists")
	}
	nodeMetrics := make(NodesMetrics, len(nos.Items))
	for i := range nos.Items {
		nodeMetrics[nos.Items[i].Name] = NodeMetrics{
			AllocatableCPU: nos.Items[i].Status.Allocatable.Cpu().MilliValue(),
			AllocatableMEM: nos.Items[i].Status.Allocatable.Memory().Value(),
		}
	}
	for i := range nmx.Items {
		if node, ok := nodeMetrics[nmx.Items[i].Name]; ok {
			node.CurrentCPU = nmx.Items[i].Usage.Cpu().MilliValue()
			node.CurrentMEM = nmx.Items[i].Usage.Memory().Value()
			nodeMetrics[nmx.Items[i].Name] = node
		}
	}

	var ccpu, cmem, tcpu, tmem int64
	for _, mx := range nodeMetrics {
		ccpu += mx.CurrentCPU
		cmem += mx.CurrentMEM
		tcpu += mx.AllocatableCPU
		tmem += mx.AllocatableMEM
	}
	mx.PercCPU, mx.PercMEM = ToPercentage(ccpu, tcpu), ToPercentage(cmem, tmem)

	return nil
}

func (m *MetricsServer) checkAccess(ns string, gvr *GVR, msg string) error {
	if !m.HasMetrics() {
		return errors.New("no metrics-server detected on cluster")
	}

	auth, err := m.CanI(ns, gvr, "", ListAccess)
	if err != nil {
		return err
	}
	if !auth {
		return errors.New(msg)
	}
	return nil
}

// NodesMetrics retrieves metrics for a given set of nodes.
func (*MetricsServer) NodesMetrics(nodes *v1.NodeList, metrics *mv1beta1.NodeMetricsList, mmx NodesMetrics) {
	if nodes == nil || metrics == nil {
		return
	}

	for i := range nodes.Items {
		mmx[nodes.Items[i].Name] = NodeMetrics{
			AllocatableCPU:       nodes.Items[i].Status.Allocatable.Cpu().MilliValue(),
			AllocatableMEM:       ToMB(nodes.Items[i].Status.Allocatable.Memory().Value()),
			AllocatableEphemeral: ToMB(nodes.Items[i].Status.Allocatable.StorageEphemeral().Value()),
			TotalCPU:             nodes.Items[i].Status.Capacity.Cpu().MilliValue(),
			TotalMEM:             ToMB(nodes.Items[i].Status.Capacity.Memory().Value()),
			TotalEphemeral:       ToMB(nodes.Items[i].Status.Capacity.StorageEphemeral().Value()),
		}
	}
	for i := range metrics.Items {
		mx, ok := mmx[metrics.Items[i].Name]
		if !ok {
			continue
		}
		mx.CurrentCPU = metrics.Items[i].Usage.Cpu().MilliValue()
		mx.CurrentMEM = ToMB(metrics.Items[i].Usage.Memory().Value())
		mx.AvailableCPU = mx.AllocatableCPU - mx.CurrentCPU
		mx.AvailableMEM = mx.AllocatableMEM - mx.CurrentMEM
		mmx[metrics.Items[i].Name] = mx
	}
}

// FetchNodesMetricsMap fetch node metrics as a map.
func (m *MetricsServer) FetchNodesMetricsMap(ctx context.Context) (NodesMetricsMap, error) {
	mm, err := m.FetchNodesMetrics(ctx)
	if err != nil {
		return nil, err
	}

	hh := make(NodesMetricsMap, len(mm.Items))
	for i := range mm.Items {
		mx := mm.Items[i]
		hh[mx.Name] = &mx
	}

	return hh, nil
}

// FetchNodesMetrics return all metrics for nodes.
func (m *MetricsServer) FetchNodesMetrics(ctx context.Context) (*mv1beta1.NodeMetricsList, error) {
	const msg = "user is not authorized to list node metrics"

	mx := new(mv1beta1.NodeMetricsList)
	if err := m.checkAccess(ClusterScope, NmxGVR, msg); err != nil {
		return mx, err
	}

	const key = "nodes"
	entry, err := m.fetchList(ctx, key, func(ctx context.Context) (any, error) {
		return m.listNodesMetrics(ctx)
	})
	if err != nil {
		return mx, err
	}
	mxList, ok := entry.(*mv1beta1.NodeMetricsList)
	if !ok {
		return nil, fmt.Errorf("expected nodemetricslist but got %T", entry)
	}

	return mxList, nil
}

func (m *MetricsServer) listNodesMetrics(ctx context.Context) (*mv1beta1.NodeMetricsList, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}
	mxList := &mv1beta1.NodeMetricsList{
		Items: make([]mv1beta1.NodeMetrics, 0, metricsPageSize),
	}
	opts := metav1.ListOptions{Limit: metricsPageSize}
	for {
		page, err := client.MetricsV1beta1().NodeMetricses().List(ctx, opts)
		if err != nil {
			return nil, err
		}
		mxList.Items = append(mxList.Items, page.Items...)
		if page.Continue == "" {
			break
		}
		opts.Continue = page.Continue
	}

	return mxList, nil
}

// FetchNodeMetrics return all metrics for nodes.
func (m *MetricsServer) FetchNodeMetrics(ctx context.Context, n string) (*mv1beta1.NodeMetrics, error) {
	const msg = "user is not authorized to list node metrics"

	mx := new(mv1beta1.NodeMetrics)
	if err := m.checkAccess(ClusterScope, NmxGVR, msg); err != nil {
		return mx, err
	}

	mmx, err := m.FetchNodesMetricsMap(ctx)
	if err != nil {
		return nil, err
	}

	mx, ok := mmx[n]
	if !ok {
		return nil, fmt.Errorf("unable to retrieve node metrics for %q", n)
	}
	return mx, nil
}

// FetchPodsMetricsMap fetch pods metrics as a map.
func (m *MetricsServer) FetchPodsMetricsMap(ctx context.Context, ns string) (PodsMetricsMap, error) {
	mm, err := m.FetchPodsMetrics(ctx, ns)
	if err != nil {
		return nil, err
	}

	hh := make(PodsMetricsMap, len(mm.Items))
	for i := range mm.Items {
		mx := mm.Items[i]
		hh[FQN(mx.Namespace, mx.Name)] = &mx
	}

	return hh, nil
}

// FetchPodsMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodsMetrics(ctx context.Context, ns string) (*mv1beta1.PodMetricsList, error) {
	mx := new(mv1beta1.PodMetricsList)
	const msg = "user is not authorized to list pods metrics"

	if ns == NamespaceAll {
		ns = BlankNamespace
	}
	if err := m.checkAccess(ns, PmxGVR, msg); err != nil {
		return mx, err
	}

	key := FQN(ns, "pods")
	entry, err := m.fetchList(ctx, key, func(ctx context.Context) (any, error) {
		return m.listPodsMetrics(ctx, ns)
	})
	if err != nil {
		return mx, err
	}
	mxList, ok := entry.(*mv1beta1.PodMetricsList)
	if !ok {
		return mx, fmt.Errorf("expected PodMetricsList but got %T", entry)
	}

	return mxList, nil
}

func (m *MetricsServer) listPodsMetrics(ctx context.Context, ns string) (*mv1beta1.PodMetricsList, error) {
	client, err := m.MXDial()
	if err != nil {
		return nil, err
	}
	mxList := &mv1beta1.PodMetricsList{
		Items: make([]mv1beta1.PodMetrics, 0, metricsPageSize),
	}
	opts := metav1.ListOptions{Limit: metricsPageSize}
	for {
		page, err := client.MetricsV1beta1().PodMetricses(ns).List(ctx, opts)
		if err != nil {
			return nil, err
		}
		mxList.Items = append(mxList.Items, page.Items...)
		if page.Continue == "" {
			break
		}
		opts.Continue = page.Continue
	}

	return mxList, nil
}

// FetchContainersMetrics returns a pod's containers metrics.
func (m *MetricsServer) FetchContainersMetrics(ctx context.Context, fqn string) (ContainersMetrics, error) {
	mm, err := m.FetchPodMetrics(ctx, fqn)
	if err != nil {
		return nil, err
	}

	cmx := make(ContainersMetrics, len(mm.Containers))
	for i := range mm.Containers {
		c := mm.Containers[i]
		cmx[c.Name] = &c
	}

	return cmx, nil
}

// FetchPodMetrics return all metrics for pods in a given namespace.
func (m *MetricsServer) FetchPodMetrics(ctx context.Context, fqn string) (*mv1beta1.PodMetrics, error) {
	var mx *mv1beta1.PodMetrics
	const msg = "user is not authorized to list pod metrics"

	ns, _ := Namespaced(fqn)
	if ns == NamespaceAll {
		ns = BlankNamespace
	}
	if err := m.checkAccess(ns, PmxGVR, msg); err != nil {
		return mx, err
	}

	mmx, err := m.FetchPodsMetricsMap(ctx, ns)
	if err != nil {
		return nil, err
	}
	pmx, ok := mmx[fqn]
	if !ok {
		return nil, fmt.Errorf("unable to locate pod metrics for pod %q", fqn)
	}

	return pmx, nil
}

// PodsMetrics retrieves metrics for all pods in a given namespace.
func (*MetricsServer) PodsMetrics(pods *mv1beta1.PodMetricsList, mmx PodsMetrics) {
	if pods == nil {
		return
	}

	// Compute all pod's containers metrics.
	for i := range pods.Items {
		var mx PodMetrics
		for _, c := range pods.Items[i].Containers {
			mx.CurrentCPU += c.Usage.Cpu().MilliValue()
			mx.CurrentMEM += ToMB(c.Usage.Memory().Value())
		}
		mmx[pods.Items[i].Namespace+"/"+pods.Items[i].Name] = mx
	}
}

// ----------------------------------------------------------------------------
// Helpers...

// MegaByte represents a megabyte.
const MegaByte = 1024 * 1024

// ToMB converts bytes to megabytes.
func ToMB(v int64) int64 {
	return v / MegaByte
}

// ToPercentage computes percentage as string otherwise n/aa.
func ToPercentage(v, dv int64) int {
	if dv == 0 {
		return 0
	}

	return int(math.Floor((float64(v) / float64(dv)) * 100))
}

// ToPercentageStr computes percentage, but if v2 is 0, it will return NAValue instead of 0.
func ToPercentageStr(v, dv int64) string {
	if dv == 0 {
		return NA
	}

	return strconv.Itoa(ToPercentage(v, dv))
}
