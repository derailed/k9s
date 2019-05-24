package watch

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

// AllNamespaces designates all namespaces.
const AllNamespaces = ""

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

// StoreInformer an informer that allows listeners registration.
type StoreInformer interface {
	cache.SharedIndexInformer
	Get(fqn string, opts metav1.GetOptions) (interface{}, error)
	List(ns string, opts metav1.ListOptions) k8s.Collection
}

// Meta represents a collection of cluster wide watchers.
type Meta struct {
	informers   map[string]StoreInformer
	client      k8s.Connection
	podInformer *Pod
	listenerFn  TableListenerFn
}

// NewMeta creates a new cluster resource informer
func NewMeta(client k8s.Connection, ns string) *Meta {
	m := Meta{client: client, informers: map[string]StoreInformer{}}

	nsAccess := m.client.CanIAccess("", "", "namespaces", []string{"list", "watch"})
	ns, err := client.Config().CurrentNamespaceName()
	// User did not lock NS. Check all ns access if not bail
	if err != nil && !nsAccess {
		log.Panic().Msg("Unauthorized access to list namespaces. Please specify a namespace")
	}

	// Namespace is locks in check if user has auth for this  ns access.
	if ns != AllNamespaces && !nsAccess {
		if !m.client.CanIAccess("", ns, "namespaces", []string{"get", "watch"}) {
			log.Panic().Msgf("Unauthorized access to namespace %q", ns)
		}
		m.init(ns)
	} else {
		m.init(AllNamespaces)
	}

	return &m
}

func (m *Meta) init(ns string) {
	po := NewPod(m.client, ns)

	m.informers = map[string]StoreInformer{
		NodeIndex:      NewNode(m.client),
		PodIndex:       po,
		ContainerIndex: NewContainer(po),
	}

	if m.client.HasMetrics() {
		if m.client.CanIAccess("", ns, "metrics.k8s.io", []string{"list", "watch"}) {
			m.informers[NodeMXIndex] = NewNodeMetrics(m.client)
			m.informers[PodMXIndex] = NewPodMetrics(m.client, ns)
		}
	}
}

// CheckAccess checks if current user as enought RBAC fu to access watched resources.
func (m *Meta) checkAccess(ns string) error {
	if !m.client.CanIAccess(ns, "nodes", "nodes", []string{"list", "watch"}) {
		return fmt.Errorf("Not authorized to list/watch nodes")
	}
	if !m.client.CanIAccess(ns, "pods", "pods", []string{"list", "watch"}) {
		return fmt.Errorf("Not authorized to list/watch pods in namespace %s", ns)
	}

	return nil
}

// List items from store.
func (m *Meta) List(res, ns string, opts metav1.ListOptions) (k8s.Collection, error) {
	if m == nil {
		return nil, fmt.Errorf("No meta exists")
	}
	if i, ok := m.informers[res]; ok {
		return i.List(ns, opts), nil
	}

	return nil, fmt.Errorf("No informer found for resource %s:%q", res, ns)
}

// Get a resource by name.
func (m Meta) Get(res, fqn string, opts metav1.GetOptions) (interface{}, error) {
	if informer, ok := m.informers[res]; ok {
		return informer.Get(fqn, opts)
	}

	return nil, fmt.Errorf("No informer found for resource %s:%q", res, fqn)
}

// Run starts watching cluster resources.
func (m *Meta) Run(closeCh <-chan struct{}) {
	for i := range m.informers {
		go func(informer StoreInformer, c <-chan struct{}) {
			informer.Run(c)
		}(m.informers[i], closeCh)
	}
}
