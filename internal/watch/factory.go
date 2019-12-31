package watch

import (
	"fmt"
	"time"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	di "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
)

const (
	defaultResync = 10 * time.Minute
	allNamespaces = ""
	clusterScope  = "-"
)

// Factory tracks various resource informers.
type Factory struct {
	factories  map[string]di.DynamicSharedInformerFactory
	client     client.Connection
	stopChan   chan struct{}
	forwarders Forwarders
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
	f.stopChan = make(chan struct{})
	for ns, fac := range f.factories {
		log.Debug().Msgf("Starting factory in ns %q", ns)
		fac.Start(f.stopChan)
	}
}

// Terminate terminates all watchers and forwards.
func (f *Factory) Terminate() {
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
func (f *Factory) List(gvr, ns string, sel labels.Selector) ([]runtime.Object, error) {
	inf, err := f.CanForResource(ns, gvr, "list")
	if err != nil {
		return nil, err
	}
	if ns == clusterScope {
		return inf.Lister().List(sel)
	}

	return inf.Lister().ByNamespace(ns).List(sel)
}

// Get retrieves a given resource.
func (f *Factory) Get(gvr, path string, sel labels.Selector) (runtime.Object, error) {
	ns, n := namespaced(path)
	inf, err := f.CanForResource(ns, gvr, "get")
	if err != nil {
		return nil, err
	}
	if ns == clusterScope {
		return inf.Lister().Get(n)
	}

	return inf.Lister().ByNamespace(ns).Get(n)
}

// WaitForCacheSync waits for all factories to update their cache.
func (f *Factory) WaitForCacheSync() {
	for _, fac := range f.factories {
		m := fac.WaitForCacheSync(f.stopChan)
		for k, v := range m {
			log.Debug().Msgf("CACHE -- Loaded %q:%v", k, v)
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

// SetActive sets the active namespace.
// BOZO!! Check ns access for resource??
func (f *Factory) SetActiveNS(ns string) {
	if !f.isClusterWide() {
		f.ensureFactory(ns)
	}
}

func (f *Factory) isClusterWide() bool {
	_, ok := f.factories[allNamespaces]
	return ok
}

func (f *Factory) preload(_ string) {
	verbs := []string{"get", "list", "watch"}
	_, _ = f.CanForResource("", "apiextensions.k8s.io/v1beta1/customresourcedefinitions", verbs...)
	// BOZO!!
	// _, _ = f.CanForResource(ns, "v1/pods", verbs...)
	// _, _ = f.CanForResource(clusterScope, "rbac.authorization.k8s.io/v1/clusterroles", verbs...)
	// _, _ = f.CanForResource(allNamespaces, "rbac.authorization.k8s.io/v1/roles", verbs...)
}

// CanForResource return an informer is user has access.
func (f *Factory) CanForResource(ns, gvr string, verbs ...string) (informers.GenericInformer, error) {
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
	fact.Start(f.stopChan)

	return inf
}

func (f *Factory) ensureFactory(ns string) di.DynamicSharedInformerFactory {
	if f.isClusterWide() {
		ns = allNamespaces
	}
	if fac, ok := f.factories[ns]; ok {
		return fac
	}

	f.factories[ns] = di.NewFilteredDynamicSharedInformerFactory(
		f.client.DynDialOrDie(),
		defaultResync,
		ns,
		nil,
	)
	f.preload(ns)

	return f.factories[ns]
}

// AddForwarder registers a new portforward for a given container.
func (f *Factory) AddForwarder(pf Forwarder) {
	f.forwarders[pf.Path()] = pf
}

// DeleteForwarder deletes portforward for a given container.
func (f *Factory) DeleteForwarder(path string) {
	fwd, ok := f.forwarders[path]
	if !ok {
		return
	}
	fwd.Stop()
	delete(f.forwarders, path)
}

// Forwarders returns all portforwards.
func (f *Factory) Forwarders() Forwarders {
	return f.forwarders
}

// ForwarderFor returns a portforward for a given container or nil if none exists.
func (f *Factory) ForwarderFor(path string) (Forwarder, bool) {
	fwd, ok := f.forwarders[path]
	return fwd, ok
}
