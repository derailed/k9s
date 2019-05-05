package watch

import (
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
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// PodMXIndex track store indexer.
	PodMXIndex string = "pmx"
)

// PodMetrics tracks pod metrics.
type PodMetrics struct {
	cache.SharedIndexInformer

	client k8s.Connection
	ns     string
}

// NewPodMetrics returns a pod metrics informer.
func NewPodMetrics(client k8s.Connection, ns string) *PodMetrics {
	mx := PodMetrics{
		ns:     ns,
		client: client,
	}

	if client == nil {
		return &mx
	}

	c, err := client.MXDial()
	if err != nil {
		return &mx
	}
	mx.SharedIndexInformer = NewPodMetricsInformer(c, ns, 0, cache.Indexers{})

	return &mx
}

// List pod metrics from store.
func (p *PodMetrics) List(ns string) (k8s.Collection, error) {
	var res k8s.Collection
	for _, o := range p.GetStore().List() {
		mx := o.(*mv1beta1.PodMetrics)
		if ns == "" || mx.Namespace == ns {
			res = append(res, mx)
		}
	}
	return res, nil
}

// Get pod metrics from store.
func (p *PodMetrics) Get(fqn string) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(fqn)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("PodMetric for %q not found", fqn)
	}

	return o, nil
}

// SetListener register an event listiner.
func (p *PodMetrics) SetListener(ns string, cb TableListenerFn) {}

// UnsetListener unregister event listener.
func (p *PodMetrics) UnsetListener(ns string) {}

// OnAdd notify pod added.
func (p *PodMetrics) OnAdd(obj interface{}) {
	po := obj.(*mv1beta1.PodMetrics)
	fqn := MetaFQN(po.ObjectMeta)
	log.Debug().Msgf("MX Added %s", fqn)
}

// OnUpdate notify pod updated.
func (p *PodMetrics) OnUpdate(oldObj, newObj interface{}) {
	opo, npo := oldObj.(*mv1beta1.PodMetrics), newObj.(*mv1beta1.PodMetrics)

	k1 := MetaFQN(opo.ObjectMeta)
	k2 := MetaFQN(npo.ObjectMeta)
	log.Debug().Msgf("MX Updated %#v -- %#v", k1, k2)
}

// OnDelete notify pod was deleted.
func (p *PodMetrics) OnDelete(obj interface{}) {
	po := obj.(*mv1beta1.PodMetrics)
	key := MetaFQN(po.ObjectMeta)
	log.Debug().Msgf("MX Delete %s", key)
}

// NewPodMetricsInformer return an informer to return pod metrix.
func NewPodMetricsInformer(client *versioned.Clientset, ns string, sync time.Duration, idxs cache.Indexers) cache.SharedIndexInformer {
	pw := NewPodMxWatcher(client, ns)

	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				l, err := client.MetricsV1beta1().PodMetricses(ns).List(opts)
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

const podMXRefresh = 15 * time.Second

// PodMxWatcher tracks pod metrics.
type PodMxWatcher struct {
	client    *versioned.Clientset
	ns        string
	cache     map[string]runtime.Object
	eventChan chan watch.Event
	doneChan  chan struct{}
}

// NewPodMxWatcher returns a new metrics watcher.
func NewPodMxWatcher(c *versioned.Clientset, ns string) *PodMxWatcher {
	return &PodMxWatcher{
		client:    c,
		ns:        ns,
		eventChan: make(chan watch.Event),
		doneChan:  make(chan struct{}),
		cache:     map[string]runtime.Object{},
	}
}

// Run watcher to monitor pod metrics.
func (p *PodMxWatcher) Run() {
	go func() {
		defer log.Debug().Msg("Podmetrics watcher canceled!")
		for {
			select {
			case <-p.doneChan:
				return
			case <-time.After(podMXRefresh):
				list, err := p.client.MetricsV1beta1().PodMetricses(p.ns).List(metav1.ListOptions{})
				if err != nil {
					log.Error().Err(err).Msg("Fetch pod metrics")
				}
				p.update(list, true)
			}
		}
	}()
}

func (p *PodMxWatcher) update(list *mv1beta1.PodMetricsList, notify bool) {
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
func (p *PodMxWatcher) Stop() {
	log.Debug().Msg("Stopping pod watcher!")
	close(p.doneChan)
	close(p.eventChan)
}

// ResultChan retrieves event channel.
func (p *PodMxWatcher) ResultChan() <-chan watch.Event {
	return p.eventChan
}

func (p *PodMxWatcher) deltas(m1, m2 *mv1beta1.PodMetrics) bool {
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
			log.Debug().Msgf("Missing container %s", k2)
			return true
		}
		if resourceDiff(v1, v2) {
			log.Debug().Msgf("Resources mismatch on container %s", k2)
			return true
		}
	}

	return false
}

// ----------------------------------------------------------------------------
// Helpers...

func resourceDiff(l1, l2 v1.ResourceList) bool {
	c1, c2 := l1[v1.ResourceCPU], l2[v1.ResourceCPU]
	if c1.Cmp(c2) != 0 {
		log.Debug().Msgf("CPU Delta %v vs %v", c1, c2)
		return true
	}
	m1, m2 := l1[v1.ResourceMemory], l2[v1.ResourceMemory]
	if m1.Cmp(m2) != 0 {
		log.Debug().Msgf("MEM Delta %d vs %d", m1.Value(), m2.Value())
		return true
	}
	return false
}
