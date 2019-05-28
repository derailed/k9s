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
)

const (
	// NodeMXIndex track store indexer.
	NodeMXIndex string = "nmx"
	// NodeMXRefresh node metrics sync rate.
	nodeMXRefresh = 30 * time.Second
)

// NodeMetrics tracks node metrics.
type NodeMetrics struct {
	cache.SharedIndexInformer

	client k8s.Connection
}

// NewNodeMetrics returns a node metrics informer.
func NewNodeMetrics(c k8s.Connection) *NodeMetrics {
	return &NodeMetrics{
		SharedIndexInformer: newNodeMetricsInformer(c, 0, cache.Indexers{}),
		client:              c,
	}
}

// List node metrics from store.
func (p *NodeMetrics) List(_ string, opts metav1.ListOptions) k8s.Collection {
	return p.GetStore().List()
}

// Get node metrics from store.
func (p *NodeMetrics) Get(MetaFQN string, opts metav1.GetOptions) (interface{}, error) {
	o, ok, err := p.GetStore().GetByKey(MetaFQN)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("No node metrics for %q", MetaFQN)
	}

	return o, nil
}

// NewNodeMetricsInformer return an informer to return node metrix.
func newNodeMetricsInformer(client k8s.Connection, sync time.Duration, idxs cache.Indexers) cache.SharedIndexInformer {
	pw := newNodeMxWatcher(client)

	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				c, err := client.MXDial()
				if err != nil {
					return nil, err
				}
				l, err := c.MetricsV1beta1().NodeMetricses().List(opts)
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

// nodeMxWatcher tracks node metrics.
type nodeMxWatcher struct {
	client    k8s.Connection
	cache     map[string]runtime.Object
	eventChan chan watch.Event
	doneChan  chan struct{}
}

// NewnodeMxWatcher returns a new metrics watcher.
func newNodeMxWatcher(c k8s.Connection) *nodeMxWatcher {
	return &nodeMxWatcher{
		client:    c,
		cache:     map[string]runtime.Object{},
		eventChan: make(chan watch.Event),
		doneChan:  make(chan struct{}),
	}
}

// Run watcher to monitor node metrics.
func (n *nodeMxWatcher) Run() {
	go func() {
		defer log.Debug().Msg("Node metrics watcher canceled!")
		for {
			select {
			case <-n.doneChan:
				return
			case <-time.After(nodeMXRefresh):
				c, err := n.client.MXDial()
				if err != nil {
					return
				}
				list, err := c.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
				if err != nil {
					log.Error().Err(err).Msg("Fetch node metrics")
				}
				n.update(list, true)
			}
		}
	}()
}

// Stop the metrics informer.
func (n *nodeMxWatcher) Stop() {
	log.Debug().Msg("Stopping node watcher!")
	close(n.doneChan)
	close(n.eventChan)
}

// ResultChan retrieves event channel.
func (n *nodeMxWatcher) ResultChan() <-chan watch.Event {
	return n.eventChan
}

func (n *nodeMxWatcher) update(list *mv1beta1.NodeMetricsList, notify bool) {
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
			if !resourceDiff(v1.(*mv1beta1.NodeMetrics).Usage, v.(*mv1beta1.NodeMetrics).Usage) {
				continue
			}
			kind = watch.Modified
		}
		if notify {
			n.eventChan <- watch.Event{Type: kind, Object: v}
		}
		n.cache[k] = v
	}
}
