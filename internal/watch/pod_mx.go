package watch

import (
	"errors"
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

const (
	// PodMXIndex track store indexer.
	PodMXIndex string = "pmx"
	// PodMXRefresh pod metrics sync rate.
	podMXRefresh = 15 * time.Second
)

// PodMetrics tracks pod metrics.
type PodMetrics struct {
	cache.SharedIndexInformer
	client k8s.Connection
	ns     string
}

// NewPodMetrics returns a pod metrics informer.
func NewPodMetrics(c k8s.Connection, ns string) *PodMetrics {
	return &PodMetrics{
		SharedIndexInformer: newPodMetricsInformer(c, ns, 0, cache.Indexers{}),
		ns:                  ns,
		client:              c,
	}
}

// List pod metrics from store.
func (p *PodMetrics) List(ns string) k8s.Collection {
	var res k8s.Collection
	for _, o := range p.GetStore().List() {
		mx := o.(*mv1beta1.PodMetrics)
		if ns == "" || mx.Namespace == ns {
			res = append(res, mx)
		}
	}

	return res
}

// Get pod metrics from store.
func (p *PodMetrics) Get(fqn string) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("No pod metrics for %q found", fqn)
	}

	return o, nil
}

// NewPodMetricsInformer return an informer to return pod metrix.
func newPodMetricsInformer(client k8s.Connection, ns string, sync time.Duration, idxs cache.Indexers) cache.SharedIndexInformer {
	pw := newPodMxWatcher(client, ns)

	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				c, err := client.MXDial()
				if err != nil {
					return nil, err
				}

				if !client.HasMetrics() {
					return nil, errors.New("metrics-server not supported")
				}

				l, err := c.MetricsV1beta1().PodMetricses(ns).List(opts)
				if err == nil {
					pw.update(l, false)
				}
				return l, err
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				pw.Run()
				return pw, nil
			},
		},
		&mv1beta1.PodMetrics{},
		sync,
		idxs,
	)
}

// PodMxWatcher tracks pod metrics.
type podMxWatcher struct {
	client    k8s.Connection
	ns        string
	cache     map[string]runtime.Object
	eventChan chan watch.Event
	doneChan  chan struct{}
}

// NewpodMxWatcher returns a new metrics watcher.
func newPodMxWatcher(c k8s.Connection, ns string) *podMxWatcher {
	return &podMxWatcher{
		client:    c,
		ns:        ns,
		eventChan: make(chan watch.Event),
		doneChan:  make(chan struct{}),
		cache:     map[string]runtime.Object{},
	}
}

// Run watcher to monitor pod metrics.
func (p *podMxWatcher) Run() {
	go func() {
		defer log.Debug().Msg("Podmetrics watcher canceled!")
		for {
			select {
			case <-p.doneChan:
				return
			case <-time.After(podMXRefresh):
				c, err := p.client.MXDial()
				if err != nil || !p.client.HasMetrics() {
					return
				}

				list, err := c.MetricsV1beta1().PodMetricses(p.ns).List(metav1.ListOptions{})
				if err != nil {
					log.Error().Err(err).Msg("Fetch pod metrics")
				}
				p.update(list, true)
			}
		}
	}()
}

func (p *podMxWatcher) update(list *mv1beta1.PodMetricsList, notify bool) {
	fqns := map[string]runtime.Object{}
	for i := range list.Items {
		fqn := MetaFQN(list.Items[i].ObjectMeta)
		fqns[fqn] = &list.Items[i]
	}

	for k, v := range p.cache {
		if _, ok := fqns[k]; !ok {
			if notify {
				p.eventChan <- watch.Event{
					Type:   watch.Deleted,
					Object: v,
				}
			}
			delete(p.cache, k)
		}
	}

	for k, v := range fqns {
		kind := watch.Added
		if v1, ok := p.cache[k]; ok {
			if !p.deltas(v1.(*mv1beta1.PodMetrics), v.(*mv1beta1.PodMetrics)) {
				continue
			}
			kind = watch.Modified
		}

		if notify {
			p.eventChan <- watch.Event{
				Type:   kind,
				Object: v,
			}
		}
		p.cache[k] = v
	}
}

// Stop the metrics informer.
func (p *podMxWatcher) Stop() {
	log.Debug().Msg("Stopping pod watcher!")
	close(p.doneChan)
	close(p.eventChan)
}

// ResultChan retrieves event channel.
func (p *podMxWatcher) ResultChan() <-chan watch.Event {
	return p.eventChan
}

func (p *podMxWatcher) deltas(m1, m2 *mv1beta1.PodMetrics) bool {
	mm1 := map[string]v1.ResourceList{}
	for _, co := range m1.Containers {
		mm1[co.Name] = co.Usage
	}
	mm2 := map[string]v1.ResourceList{}
	for _, co := range m2.Containers {
		mm2[co.Name] = co.Usage
	}

	for k2, v2 := range mm2 {
		v1, ok := mm1[k2]
		if !ok {
			return true
		}
		if resourceDiff(v1, v2) {
			return true
		}
	}

	return false
}
