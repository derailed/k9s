package watch

import (
	"errors"
	"fmt"
	"sync"

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

// Informer represents a collection of cluster wide watchers.
type Informer struct {
	informers   map[string]StoreInformer
	client      k8s.Connection
	podInformer *Pod
	listenerFn  TableListenerFn
	initOnce    sync.Once
}

// NewInformer creates a new cluster resource informer
func NewInformer(client k8s.Connection, ns string) *Informer {
	log.Debug().Msgf(">> Starting Informer")
	i := Informer{client: client, informers: map[string]StoreInformer{}}

	nsAccess := i.client.CanIAccess("", "", "namespaces", []string{"list", "watch"})
	ns, err := client.Config().CurrentNamespaceName()
	// User did not lock NS. Check all ns access if not bail
	if err != nil && !nsAccess {
		log.Panic().Msg("Unauthorized access to list namespaces. Please specify a namespace")
	}

	// Namespace is locked in. check if user has auth for this ns access.
	if ns != AllNamespaces && !nsAccess {
		if !i.client.CanIAccess("", ns, "namespaces", []string{"get", "watch"}) {
			log.Panic().Msgf("Unauthorized access to namespace %q", ns)
		}
		i.init(ns)
	} else {
		i.init(AllNamespaces)
	}

	return &i
}

func (i *Informer) init(ns string) {
	i.initOnce.Do(func() {
		po := NewPod(i.client, ns)
		i.informers = map[string]StoreInformer{
			NodeIndex:      NewNode(i.client),
			PodIndex:       po,
			ContainerIndex: NewContainer(po),
		}

		if !i.client.HasMetrics() {
			return
		}

		if i.client.CanIAccess("", ns, "metrics.k8s.io", []string{"list", "watch"}) {
			i.informers[NodeMXIndex] = NewNodeMetrics(i.client)
			i.informers[PodMXIndex] = NewPodMetrics(i.client, ns)
		}
	})
}

// List items from store.
func (i *Informer) List(res, ns string, opts metav1.ListOptions) (k8s.Collection, error) {
	if i == nil {
		return nil, errors.New("Invalid informer")
	}

	if i, ok := i.informers[res]; ok {
		return i.List(ns, opts), nil
	}

	return nil, fmt.Errorf("No informer found for resource %s:%q", res, ns)
}

// Get a resource by name.
func (i Informer) Get(res, fqn string, opts metav1.GetOptions) (interface{}, error) {
	if informer, ok := i.informers[res]; ok {
		return informer.Get(fqn, opts)
	}

	return nil, fmt.Errorf("No informer found for resource %s:%q", res, fqn)
}

// Run starts watching cluster resources.
func (i *Informer) Run(closeCh <-chan struct{}) {
	for name := range i.informers {
		go func(si StoreInformer, c <-chan struct{}) {
			si.Run(c)
		}(i.informers[name], closeCh)
	}
}
