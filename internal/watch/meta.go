package watch

import (
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

type (
	// Row represents a collection of string fields.
	Row []string

	// RowEvent represents a call for action after a resource reconciliation.
	// Tracks whether a resource got added, deleted or updated.
	RowEvent struct {
		Action watch.EventType
		Fields Row
		Deltas Row
	}

	// RowEvents tracks resource update events.
	RowEvents map[string]*RowEvent

	// TableData tracks a K8s resource for tabular display.
	TableData struct {
		Header    Row
		Rows      RowEvents
		Namespace string
	}
)

// TableListenerFn represents a table data listener.
type TableListenerFn func(TableData)

// CallbackInformer an informer that allows listeners registration.
type CallbackInformer interface {
	cache.SharedIndexInformer
	Get(fqn string) (interface{}, error)
	List(ns string) (k8s.Collection, error)
	SetListener(ns string, cb TableListenerFn)
	UnsetListener(ns string)
}

// Meta represents a collection of cluster wide watchers.
type Meta struct {
	informers   map[string]CallbackInformer
	client      k8s.Connection
	podInformer *Pod
	listenerFn  TableListenerFn
}

// NewMeta creates a new cluster resource informer
func NewMeta(client k8s.Connection, ns string) *Meta {
	m := Meta{client: client, informers: map[string]CallbackInformer{}}
	m.init(ns)

	return &m
}

func (m *Meta) init(ns string) {
	po := NewPod(m.client, ns)
	m.informers = map[string]CallbackInformer{
		NodeIndex:      NewNode(m.client),
		NodeMXIndex:    NewNodeMetrics(m.client),
		PodIndex:       po,
		PodMXIndex:     NewPodMetrics(m.client, ns),
		ContainerIndex: NewContainer(po),
	}
}

// List items from store.
func (m *Meta) List(res, ns string) (k8s.Collection, error) {
	if m == nil {
		return nil, fmt.Errorf("No meta exists")
	}
	if i, ok := m.informers[res]; ok {
		return i.List(ns)
	}

	return nil, fmt.Errorf("No informer found for resource %s", res)
}

// RegisterListener register table data listeners.
func (m *Meta) RegisterListener(res, ns string, cb TableListenerFn) {
	if informer, ok := m.informers[res]; ok {
		informer.SetListener(ns, cb)
	} else {
		log.Error().Msgf("No informer for %q:%s", ns, res)
	}
}

// UnregisterListener register table data listeners.
func (m *Meta) UnregisterListener(ns string) {
	for _, i := range m.informers {
		i.UnsetListener(ns)
	}
}

// Get a resource by name.
func (m Meta) Get(res, fqn string) (interface{}, error) {
	if informer, ok := m.informers[res]; ok {
		return informer.Get(fqn)
	}

	return nil, errors.New("Unable to local resource")
}

// Run starts watching cluster resources.
func (m *Meta) Run(closeCh <-chan struct{}) {
	for i := range m.informers {
		go func(informer CallbackInformer, c <-chan struct{}) {
			informer.Run(c)
		}(m.informers[i], closeCh)
	}
}
