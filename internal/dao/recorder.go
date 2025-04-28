package dao

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/slogs"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/cache"
)

var MxRecorder *Recorder

const (
	seriesCacheSize   = 600
	seriesCacheExpiry = 3 * time.Hour
	seriesRecordRate  = 1 * time.Minute
	nodeMetrics       = "node"
	podMetrics        = "pod"
)

type MetricsChan chan TimeSeries

type TimeSeries []Point

type Point struct {
	Time  time.Time
	Tags  map[string]string
	Value client.NodeMetrics
}

type Recorder struct {
	conn   client.Connection
	series *cache.LRUExpireCache
	mxChan MetricsChan
	mx     sync.RWMutex
}

func DialRecorder(c client.Connection) *Recorder {
	if MxRecorder != nil {
		return MxRecorder
	}
	MxRecorder = &Recorder{
		conn:   c,
		series: cache.NewLRUExpireCache(seriesCacheSize),
	}

	return MxRecorder
}

func ResetRecorder(c client.Connection) {
	MxRecorder = nil
	DialRecorder(c)
}

func (r *Recorder) Clear() {
	r.mx.Lock()
	defer r.mx.Unlock()

	kk := r.series.Keys()
	for _, k := range kk {
		r.series.Remove(k)
	}
}

func (r *Recorder) dispatchSeries(kind, ns string) {
	if r.mxChan == nil {
		return
	}
	kk := r.series.Keys()
	hour := time.Now().Add(-1 * time.Hour)
	ts := make(TimeSeries, 0, len(kk))
	for _, k := range kk {
		if v, ok := r.series.Get(k); ok {
			if pt, cool := v.(Point); cool {
				if pt.Tags["type"] != kind || pt.Time.Sub(hour) < 0 {
					continue
				}
				switch kind {
				case nodeMetrics:
					ts = append(ts, pt)
				case podMetrics:
					if client.IsAllNamespaces(ns) || pt.Tags["namespace"] == ns {
						ts = append(ts, pt)
					}
				}
			}
		}
	}
	if len(ts) > 0 {
		r.mxChan <- ts
	}
}

func (r *Recorder) Watch(ctx context.Context, ns string) MetricsChan {
	r.mx.Lock()
	if r.mxChan != nil {
		close(r.mxChan)
		r.mxChan = nil
	}
	r.mxChan = make(MetricsChan, 2)
	r.mx.Unlock()

	go func() {
		kind := podMetrics
		if client.IsAllNamespaces(ns) {
			kind = nodeMetrics
		}
		switch kind {
		case podMetrics:
			if err := r.recordPodMetrics(ctx, ns); err != nil {
				slog.Error("Record pod metrics failed", slogs.Error, err)
			}
		case nodeMetrics:
			if err := r.recordNodeMetrics(ctx); err != nil {
				slog.Error("Record node metrics failed", slogs.Error, err)
			}
		}
		r.dispatchSeries(kind, ns)
		<-ctx.Done()
		r.mx.Lock()
		if r.mxChan != nil {
			close(r.mxChan)
			r.mxChan = nil
		}
		r.mx.Unlock()
	}()

	return r.mxChan
}

func (r *Recorder) Record(ctx context.Context) error {
	if err := r.recordNodeMetrics(ctx); err != nil {
		return err
	}
	return r.recordPodMetrics(ctx, client.NamespaceAll)
}

func (r *Recorder) recordNodeMetrics(ctx context.Context) error {
	f, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok {
		return errors.New("expecting factory in context")
	}
	nn, err := FetchNodes(ctx, f, "")
	if err != nil {
		return err
	}

	go func() {
		r.recordClusterMetrics(ctx, nn)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(seriesRecordRate):
				r.recordClusterMetrics(ctx, nn)
			}
		}
	}()

	return nil
}

func (r *Recorder) recordClusterMetrics(ctx context.Context, nn *v1.NodeList) {
	dial := client.DialMetrics(r.conn)
	nmx, err := dial.FetchNodesMetrics(ctx)
	if err != nil {
		slog.Error("Fetch node metrics failed", slogs.Error, err)
		return
	}

	mx := make(client.NodesMetrics, len(nn.Items))
	dial.NodesMetrics(nn, nmx, mx)
	var cmx client.NodeMetrics
	for _, m := range mx {
		cmx.CurrentCPU += m.CurrentCPU
		cmx.CurrentMEM += m.CurrentMEM
		cmx.AllocatableCPU += m.AllocatableCPU
		cmx.AllocatableMEM += m.AllocatableMEM
		cmx.TotalCPU += m.TotalCPU
		cmx.TotalMEM += m.TotalMEM
	}
	pt := Point{
		Time:  time.Now(),
		Value: cmx,
		Tags: map[string]string{
			"type": nodeMetrics,
		},
	}
	if len(nn.Items) > 0 {
		r.series.Add(pt.Time, pt, seriesCacheExpiry)
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.mxChan != nil {
		r.mxChan <- TimeSeries{pt}
	}
}

func (r *Recorder) recordPodMetrics(ctx context.Context, ns string) error {
	go func() {
		if err := r.recordPodsMetrics(ctx, ns); err != nil {
			slog.Error("Record pod metrics failed", slogs.Error, err)
		}
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(seriesRecordRate):
				// case <-time.After(5 * time.Second):
				if err := r.recordPodsMetrics(ctx, ns); err != nil {
					slog.Error("Record pod metrics failed", slogs.Error, err)
				}
			}
		}
	}()

	return nil
}

func (r *Recorder) recordPodsMetrics(ctx context.Context, ns string) error {
	f, ok := ctx.Value(internal.KeyFactory).(Factory)
	if !ok {
		return errors.New("expecting factory in context")
	}
	pp, err := FetchPods(ctx, f, ns)
	if err != nil {
		return err
	}

	pt := Point{
		Time:  time.Now(),
		Value: client.NodeMetrics{},
		Tags: map[string]string{
			"namespace": ns,
			"type":      podMetrics,
		},
	}
	dial := client.DialMetrics(r.conn)
	for i := range pp.Items {
		p := pp.Items[i]
		fqn := client.FQN(p.Namespace, p.Name)
		pmx, err := dial.FetchPodMetrics(ctx, fqn)
		if err != nil {
			continue
		}
		for _, c := range pmx.Containers {
			pt.Value.CurrentCPU += c.Usage.Cpu().MilliValue()
			pt.Value.CurrentMEM += client.ToMB(c.Usage.Memory().Value())
		}
	}
	if len(pp.Items) > 0 {
		pt.Value.AllocatableCPU = pt.Value.CurrentCPU
		pt.Value.AllocatableMEM = pt.Value.CurrentMEM
		r.series.Add(pt.Time, pt, seriesCacheExpiry)
		r.mx.Lock()
		defer r.mx.Unlock()
		if r.mxChan != nil {
			r.mxChan <- TimeSeries{pt}
		}
	}

	return nil
}

// FetchPods retrieves all pods in a given namespace.
func FetchPods(_ context.Context, f Factory, ns string) (*v1.PodList, error) {
	auth, err := f.Client().CanI(ns, client.PodGVR, "pods", []string{client.ListVerb})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("user is not authorized to list pods")
	}

	oo, err := f.List(client.PodGVR, ns, false, labels.Everything())
	if err != nil {
		return nil, err
	}
	pp := make([]v1.Pod, 0, len(oo))
	for _, o := range oo {
		var pod v1.Pod
		err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &pod)
		if err != nil {
			return nil, err
		}
		pp = append(pp, pod)
	}

	return &v1.PodList{Items: pp}, nil
}
