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

const (
	// AllNamespaces designates all namespaces.
	allNamespaces = ""
	// AllNamespaces designate the special `all` namespace.
	allNamespace = "all"
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

// StoreInformer an informer that allows listeners registration.
type StoreInformer interface {
	cache.SharedIndexInformer
	Get(fqn string, opts metav1.GetOptions) (interface{}, error)
	List(ns string, opts metav1.ListOptions) k8s.Collection
}

// Informer represents a collection of cluster wide watchers.
type Informer struct {
	Namespace   string
	informers   map[string]StoreInformer
	client      k8s.Connection
	podInformer *Pod
	listenerFn  TableListenerFn
	initOnce    sync.Once
}

// NewInformer creates a new cluster resource informer
func NewInformer(client k8s.Connection, ns string) (*Informer, error) {
	i := Informer{
		client:    client,
		Namespace: ns,
		informers: map[string]StoreInformer{},
	}
	if err := client.CheckNSAccess(ns); err != nil {
		log.Error().Err(err).Msg("Checking NS Access")
		return nil, err
	}
	i.init(ns)

	return &i, nil
}

func (i *Informer) Dump() {
	log.Debug().Msgf("Informer Dump")
	for k := range i.informers {
		log.Debug().Msgf("\t%s", k)
	}
}

func (i *Informer) init(ns string) {
	log.Debug().Msgf(">>>>> Starting Informer in namespace %q", ns)

	if ok, err := i.client.CanIAccess(ns, "pods.v1.", []string{"list", "watch"}); ok && err == nil {
		log.Debug().Msgf("Pod access granted!")
	} else {
		log.Debug().Msgf("No pod access! %t -- %#v", ok, err)
	}

	i.initOnce.Do(func() {
		po := NewPod(i.client, ns)
		i.informers = map[string]StoreInformer{
			PodIndex:       po,
			ContainerIndex: NewContainer(po),
		}
		if ok, err := i.client.CanIAccess(ns, "nodes.v1.", []string{"list", "watch"}); ok && err == nil {
			log.Debug().Msgf("CanI access nodes %t -- %#v", ok, err)
			i.informers[NodeIndex] = NewNode(i.client)
		} else {
			log.Debug().Msgf("No node access! %t -- %#v", ok, err)
		}

		if !i.client.HasMetrics() {
			return
		}

		if ok, err := i.client.CanIAccess(ns, "nodes.v1beta1.metrics.k8s.io", []string{"list", "watch"}); ok && err == nil {
			i.informers[NodeMXIndex] = NewNodeMetrics(i.client)
		} else {
			log.Debug().Msg("No node metrics access!")
		}
		if ok, err := i.client.CanIAccess(ns, "pods.v1beta1.metrics.k8s.io", []string{"list", "watch"}); ok && err == nil {
			i.informers[PodMXIndex] = NewPodMetrics(i.client, ns)
		} else {
			log.Debug().Msgf("No pod metrics access! %t -- %#v", ok, err)
		}
	})
}

// List items from store.
func (i *Informer) List(res, ns string, opts metav1.ListOptions) (k8s.Collection, error) {
	if i == nil {
		return nil, errors.New("Invalid List informer")
	}

	if i, ok := i.informers[res]; ok {
		return i.List(ns, opts), nil
	}

	return nil, fmt.Errorf("No informer found for resource %s in namespace %q", res, ns)
}

// Get a resource by name.
func (i *Informer) Get(res, fqn string, opts metav1.GetOptions) (interface{}, error) {
	if i == nil {
		return nil, errors.New("Invalid Get informer")
	}

	if informer, ok := i.informers[res]; ok {
		return informer.Get(fqn, opts)
	}

	return nil, fmt.Errorf("No informer found for resource %s in namespace %q", res, fqn)
}

// Run starts watching cluster resources.
func (i *Informer) Run(closeCh <-chan struct{}) {
	for name := range i.informers {
		go func(si StoreInformer, c <-chan struct{}) {
			si.Run(c)
		}(i.informers[name], closeCh)
	}
}
