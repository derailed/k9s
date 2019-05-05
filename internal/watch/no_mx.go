package watch

import (
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
	mv1beta1 "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	// NodeMXIndex track store indexer.
	NodeMXIndex string = "nmx"
)

// NodeMetrics tracks node metrics.
type NodeMetrics struct {
	cache.SharedIndexInformer

	client k8s.Connection
}

// NewNodeMetrics returns a node metrics informer.
func NewNodeMetrics(client k8s.Connection) *NodeMetrics {
	mx := NodeMetrics{
		client: client,
	}

	if client == nil {
		return &mx
	}

	c, err := client.MXDial()
	if err != nil {
		return &mx
	}
	mx.SharedIndexInformer = NewNodeMetricsInformer(c, 0, cache.Indexers{})
	mx.SharedIndexInformer.AddEventHandler(&mx)

	return &mx
}

// List node metrics from store.
func (p *NodeMetrics) List(string) (k8s.Collection, error) {
	return p.GetStore().List(), nil
}

// Get node metrics from store.
func (p *NodeMetrics) Get(MetaFQN string) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(MetaFQN)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("NodeMetric for %q not found", MetaFQN)
	}

	return o, nil
}

// SetListener register an event listiner.
func (p *NodeMetrics) SetListener(ns string, cb TableListenerFn) {}

// UnsetListener unregister event listener.
func (p *NodeMetrics) UnsetListener(ns string) {}

// OnAdd notify node added.
func (p *NodeMetrics) OnAdd(obj interface{}) {
	po := obj.(*mv1beta1.NodeMetrics)
	fqn := MetaFQN(po.ObjectMeta)
	log.Debug().Msgf("NMX Added %s", fqn)
}

// OnUpdate notify node updated.
func (p *NodeMetrics) OnUpdate(oldObj, newObj interface{}) {
	opo, npo := oldObj.(*mv1beta1.NodeMetrics), newObj.(*mv1beta1.NodeMetrics)

	k1 := MetaFQN(opo.ObjectMeta)
	k2 := MetaFQN(npo.ObjectMeta)
	log.Debug().Msgf("NMX Updated %#v -- %#v", k1, k2)
}

// OnDelete notify node was deleted.
func (p *NodeMetrics) OnDelete(obj interface{}) {
	po := obj.(*mv1beta1.NodeMetrics)
	key := MetaFQN(po.ObjectMeta)
	log.Debug().Msgf("NMX Delete %s", key)
}

// NewNodeMetricsInformer return an informer to return node metrix.
func NewNodeMetricsInformer(client *versioned.Clientset, sync time.Duration, idxs cache.Indexers) cache.SharedIndexInformer {
	pw := NewNodeMxWatcher(client)

	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				l, err := client.MetricsV1beta1().NodeMetricses().List(opts)
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
		&mv1beta1.NodeMetrics{},
		sync,
		idxs,
	)
}

const nodeMXRefresh = 30 * time.Second

// NodeMxWatcher tracks node metrics.
type NodeMxWatcher struct {
	client    *versioned.Clientset
	cache     map[string]runtime.Object
	eventChan chan watch.Event
	doneChan  chan struct{}
}

// NewNodeMxWatcher returns a new metrics watcher.
func NewNodeMxWatcher(c *versioned.Clientset) *NodeMxWatcher {
	return &NodeMxWatcher{
		client:    c,
		cache:     map[string]runtime.Object{},
		eventChan: make(chan watch.Event),
		doneChan:  make(chan struct{}),
	}
}

// Run watcher to monitor node metrics.
func (n *NodeMxWatcher) Run() {
	go func() {
		defer log.Debug().Msg("Node metrics watcher canceled!")
		for {
			select {
			case <-n.doneChan:
				return
			case <-time.After(nodeMXRefresh):
				list, err := n.client.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
				if err != nil {
					log.Error().Err(err).Msg("Fetch node metrics")
				}
				n.update(list, true)
			}
		}
	}()
}

func (n *NodeMxWatcher) update(list *mv1beta1.NodeMetricsList, notify bool) {
	fqns := map[string]runtime.Object{}
	for i := range list.Items {
		fqn := MetaFQN(list.Items[i].ObjectMeta)
		fqns[fqn] = &list.Items[i]
	}

	for k, v := range n.cache {
		if _, ok := fqns[k]; !ok {
			if notify {
				n.eventChan <- watch.Event{
					Type:   watch.Deleted,
					Object: v,
				}
			}
			delete(n.cache, k)
		}
	}

	for k, v := range fqns {
		kind := watch.Added
		if v1, ok := n.cache[k]; ok {
			if !n.deltas(v1.(*mv1beta1.NodeMetrics), v.(*mv1beta1.NodeMetrics)) {
				continue
			}
			kind = watch.Modified
		}

		if notify {
			n.eventChan <- watch.Event{
				Type:   kind,
				Object: v,
			}
		}
		n.cache[k] = v
	}
}

// Stop the metrics informer.
func (n *NodeMxWatcher) Stop() {
	log.Debug().Msg("Stopping node watcher!")
	close(n.doneChan)
	close(n.eventChan)
}

// ResultChan retrieves event channel.
func (n *NodeMxWatcher) ResultChan() <-chan watch.Event {
	return n.eventChan
}

func (n *NodeMxWatcher) deltas(m1, m2 *mv1beta1.NodeMetrics) bool {
	return resourceDiff(m1.Usage, m2.Usage)
}
