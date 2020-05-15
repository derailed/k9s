package watch

import (
	"fmt"
	"sync"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	di "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
)

const (
	defaultResync   = 10 * time.Minute
	defaultWaitTime = 250 * time.Millisecond
)

// Factory tracks various resource informers.
type Factory struct {
	factories  map[string]di.DynamicSharedInformerFactory
	client     client.Connection
	stopChan   chan struct{}
	forwarders Forwarders
	mx         sync.RWMutex
}

// NewFactory returns a new informers factory.
func NewFactory(client client.Connection) *Factory {
	return &Factory{
		client:     client,
		factories:  make(map[string]di.DynamicSharedInformerFactory),
		forwarders: NewForwarders(),
	}
}

// Start initializes the informers until caller cancels the context.
func (f *Factory) Start(ns string) {
	f.mx.Lock()
	defer f.mx.Unlock()

	log.Debug().Msgf("Factory START with ns `%q", ns)
	f.stopChan = make(chan struct{})
	for ns, fac := range f.factories {
		log.Debug().Msgf("Starting factory in ns %q", ns)
		fac.Start(f.stopChan)
	}
}

// Terminate terminates all watchers and forwards.
func (f *Factory) Terminate() {
	f.mx.Lock()
	defer f.mx.Unlock()

	if f.stopChan != nil {
		close(f.stopChan)
		f.stopChan = nil
	}
	for k := range f.factories {
		delete(f.factories, k)
	}
	f.forwarders.DeleteAll()
}

// List returns a resource collection.
func (f *Factory) List(gvr, ns string, wait bool, labels labels.Selector) ([]runtime.Object, error) {
	inf, err := f.CanForResource(ns, gvr, client.MonitorAccess)
	if err != nil {
		return nil, err
	}
	if wait {
		f.waitForCacheSync(ns)
	}
	if client.IsClusterScoped(ns) {
		return inf.Lister().List(labels)
	}

	if client.IsAllNamespace(ns) {
		ns = client.AllNamespaces
	}
	return inf.Lister().ByNamespace(ns).List(labels)
}

// Get retrieves a given resource.
func (f *Factory) Get(gvr, path string, wait bool, sel labels.Selector) (runtime.Object, error) {
	ns, n := namespaced(path)
	inf, err := f.CanForResource(ns, gvr, []string{client.GetVerb})
	if err != nil {
		return nil, err
	}

	if wait {
		f.waitForCacheSync(ns)
	}
	if client.IsClusterScoped(ns) {
		return inf.Lister().Get(n)
	}

	return inf.Lister().ByNamespace(ns).Get(n)
}

func (f *Factory) waitForCacheSync(ns string) {
	if client.IsClusterWide(ns) {
		ns = client.AllNamespaces
	}

	if f.isClusterWide() {
		ns = client.AllNamespaces
	}

	f.mx.RLock()
	defer f.mx.RUnlock()
	fac, ok := f.factories[ns]
	if !ok {
		return
	}

	// Hang for a sec for the cache to refresh if still not done bail out!
	c := make(chan struct{})
	go func(c chan struct{}) {
		<-time.After(defaultWaitTime)
		close(c)
	}(c)
	_ = fac.WaitForCacheSync(c)
}

// WaitForCacheSync waits for all factories to update their cache.
func (f *Factory) WaitForCacheSync() {
	for ns, fac := range f.factories {
		m := fac.WaitForCacheSync(f.stopChan)
		for k, v := range m {
			log.Debug().Msgf("CACHE `%q Loaded %t:%s", ns, v, k)
		}
	}
}

// Client return the factory connection.
func (f *Factory) Client() client.Connection {
	return f.client
}

// FactoryFor returns a factory for a given namespace.
func (f *Factory) FactoryFor(ns string) di.DynamicSharedInformerFactory {
	return f.factories[ns]
}

// SetActiveNS sets the active namespace.
func (f *Factory) SetActiveNS(ns string) {
	if !f.isClusterWide() {
		f.ensureFactory(ns)
	}
}

func (f *Factory) isClusterWide() bool {
	f.mx.RLock()
	defer f.mx.RUnlock()

	_, ok := f.factories[client.AllNamespaces]
	return ok
}

// CanForResource return an informer is user has access.
func (f *Factory) CanForResource(ns, gvr string, verbs []string) (informers.GenericInformer, error) {
	auth, err := f.Client().CanI(ns, gvr, verbs)
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("%v access denied on resource %q:%q", verbs, ns, gvr)
	}

	return f.ForResource(ns, gvr), nil
}

// ForResource returns an informer for a given resource.
func (f *Factory) ForResource(ns, gvr string) informers.GenericInformer {
	fact := f.ensureFactory(ns)
	inf := fact.ForResource(toGVR(gvr))
	if inf == nil {
		log.Error().Err(fmt.Errorf("MEOW! No informer for %q:%q", ns, gvr))
		return inf
	}

	f.mx.RLock()
	defer f.mx.RUnlock()
	fact.Start(f.stopChan)

	return inf
}

func (f *Factory) ensureFactory(ns string) di.DynamicSharedInformerFactory {
	if client.IsClusterWide(ns) {
		ns = client.AllNamespaces
	}
	f.mx.Lock()
	defer f.mx.Unlock()
	if fac, ok := f.factories[ns]; ok {
		return fac
	}
	f.factories[ns] = di.NewFilteredDynamicSharedInformerFactory(
		f.client.DynDialOrDie(),
		defaultResync,
		ns,
		nil,
	)

	return f.factories[ns]
}

// AddForwarder registers a new portforward for a given container.
func (f *Factory) AddForwarder(pf Forwarder) {
	f.mx.Lock()
	defer f.mx.Unlock()

	f.forwarders[pf.Path()] = pf
}

// DeleteForwarder deletes portforward for a given container.
func (f *Factory) DeleteForwarder(path string) {
	count := f.forwarders.Kill(path)
	log.Warn().Msgf("Deleted (%d) portforward for %q", count, path)
}

// Forwarders returns all portforwards.
func (f *Factory) Forwarders() Forwarders {
	f.mx.RLock()
	defer f.mx.RUnlock()

	return f.forwarders
}

// ForwarderFor returns a portforward for a given container or nil if none exists.
func (f *Factory) ForwarderFor(path string) (Forwarder, bool) {
	f.mx.RLock()
	defer f.mx.RUnlock()

	fwd, ok := f.forwarders[path]
	return fwd, ok
}
