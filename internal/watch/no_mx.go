package watch

import (
	"errors"
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
	c, err := client.MXDial()
	if err != nil {
		log.Error().Err(err).Msg("NodeMetrix dial")
		return nil
	}

	return cache.NewSharedIndexInformer(
		&cache.ListWatch{
			ListFunc: func(opts metav1.ListOptions) (runtime.Object, error) {
				l, err := c.MetricsV1beta1().NodeMetricses().List(opts)
				if err == nil {
					pw.update(l, false)
				}
				return l, err
			},
			WatchFunc: func(opts metav1.ListOptions) (watch.Interface, error) {
				go pw.Run()
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
	defer log.Debug().Msg("NodeMetrics informer canceled!")
	c, err := n.client.MXDial()
	if err != nil {
		log.Error().Err(err).Msg("NodeMetrix Dial Failed!")
		return
	}

	for {
		select {
		case <-n.doneChan:
			close(n.eventChan)
			return
		case <-time.After(nodeMXRefresh):
			list, err := c.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{})
			if err != nil {
				log.Error().Err(err).Msg("NodeMetrics List Failed!")
			}
			n.update(list, true)
		}
	}
}

// Stop the metrics informer.
func (n *nodeMxWatcher) Stop() {
	log.Debug().Msg("Stopping NodeMetrix informer!")
	close(n.doneChan)
}

// ResultChan retrieves event channel.
func (n *nodeMxWatcher) ResultChan() <-chan watch.Event {
	return n.eventChan
}

func (n *nodeMxWatcher) notify(event watch.Event) error {
	select {
	case n.eventChan <- event:
		return nil
	case <-n.doneChan:
		return errors.New("watcher has ben closed.")
	}
}

func (n *nodeMxWatcher) update(list *mv1beta1.NodeMetricsList, notify bool) {
	fqns := map[string]runtime.Object{}
	for i := range list.Items {
		fqn := MetaFQN(list.Items[i].ObjectMeta)
		fqns[fqn] = &list.Items[i]
	}
	n.checkDeletes(fqns, notify)
	n.checkAdds(fqns, notify)
}

func (n *nodeMxWatcher) checkDeletes(m map[string]runtime.Object, notify bool) {
	for k, v := range n.cache {
		if _, ok := m[k]; ok {
			continue
		}
		delete(n.cache, k)
		if notify && n.notify(watch.Event{Type: watch.Deleted, Object: v}) != nil {
			return
		}
	}
}

func (n *nodeMxWatcher) checkAdds(m map[string]runtime.Object, notify bool) {
	for k, v := range m {
		kind := watch.Added
		if v1, ok := n.cache[k]; ok {
			if !resourceDiff(v1.(*mv1beta1.NodeMetrics).Usage, v.(*mv1beta1.NodeMetrics).Usage) {
				continue
			}
			kind = watch.Modified
		}
		n.cache[k] = v
		if notify && n.notify(watch.Event{Type: kind, Object: v}) != nil {
			return
		}
	}
}
