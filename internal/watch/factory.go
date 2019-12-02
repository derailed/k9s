package watch

import (
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	di "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/internalinterfaces"
)

const defaultResync = 10 * time.Minute

// BOZO!!
// // Authorizer checks what a user can or cannot do to a resource.
// type Authorizer interface {
// 	// CanI returns true if the user can use these actions for a given resource.
// 	CanI(ns, gvr string, verbs []string) (bool, error)
// }

// type Connection interface {
// 	Authorizer

// 	// DialOrDie dials client api.
// 	DialOrDie() kubernetes.Interface

// 	// MXDial dials metrics api.
// 	MXDial() (*versioned.Clientset, error)

// 	// DynDialOrDie dials dynamic client api.
// 	DynDialOrDie() dynamic.Interface

// 	// RestConfigOrDie return a client configuration.
// 	RestConfigOrDie() *restclient.Config

// 	// Config returns the current kubeconfig.
// 	Config() *k8s.Config

// 	// CachedDiscovery returns a cached client.
// 	CachedDiscovery() (*disk.CachedDiscoveryClient, error)

// 	// SwithContextOrDie switch to a new kube context.
// 	SwitchContextOrDie(ctx string)
// }

// Factory tracks various resource informers.
type Factory struct {
	factories        map[string]di.DynamicSharedInformerFactory
	client           k8s.Connection
	stopChan         chan struct{}
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	activeNS         string
}

// NewFactory returns a new informers factory.
func NewFactory(client k8s.Connection) *Factory {
	return &Factory{
		client:    client,
		stopChan:  make(chan struct{}),
		factories: make(map[string]di.DynamicSharedInformerFactory),
	}
}

func (f *Factory) Dump() {
	log.Debug().Msgf("----------- FACTORIES -------------")
	for ns := range f.factories {
		log.Debug().Msgf("  Factory for NS %q", ns)
	}
	log.Debug().Msgf("-----------------------------------")
}

func (f *Factory) Debug(gvr string) {
	log.Debug().Msgf("----------- DEBUG FACTORY (%s) -------------", gvr)
	inf := f.factories[render.AllNamespaces].ForResource(toGVR(gvr))
	for i, k := range inf.Informer().GetStore().ListKeys() {
		log.Debug().Msgf("%d -- %s", i, k)
	}
}

func (f *Factory) Show(ns, gvr string) {
	log.Debug().Msgf("----------- SHOW FACTORIES %q -------------", ns)
	inf := f.ForResource(ns, gvr)
	for _, k := range inf.Informer().GetStore().ListKeys() {
		log.Debug().Msgf("  Key: %s", k)
	}
}

func (f *Factory) List(ns, gvr string, sel labels.Selector) ([]runtime.Object, error) {
	auth, err := f.Client().CanI(ns, gvr, []string{"list"})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("User has insufficient access to list %s", gvr)
	}

	log.Debug().Msgf(">>>>>>>>>>>>>> FACTORY LISTING %q -- %q", ns, gvr)
	inf := f.ForResource(ns, gvr)
	if inf == nil {
		return nil, fmt.Errorf("No resource for GVR %s", gvr)
	}

	if ns == render.ClusterWide {
		return inf.Lister().List(sel)
	}
	return inf.Lister().ByNamespace(ns).List(sel)
}

func (f *Factory) Get(ns, gvr, name string, sel labels.Selector) (runtime.Object, error) {
	log.Debug().Msgf("<<<<<<<<<<<<<<<<< FACTORY GET %q --- %q:%q", gvr, ns, name)
	auth, err := f.Client().CanI(ns, gvr, []string{"get"})
	if err != nil {
		return nil, err
	}
	if !auth {
		return nil, fmt.Errorf("User has insufficient access to get %s", gvr)
	}

	fac := f.ensureFactory(ns)
	inf := fac.ForResource(toGVR(gvr))
	if inf == nil {
		return nil, fmt.Errorf("No resource for GVR %s", gvr)
	}

	if ns == render.ClusterWide {
		return inf.Lister().Get(name)
	}
	return inf.Lister().ByNamespace(ns).Get(name)
}

func (f *Factory) WaitForCacheSync() map[schema.GroupVersionResource]bool {
	r := make(map[schema.GroupVersionResource]bool)
	for n, fac := range f.factories {
		log.Debug().Msgf(">>> WAITING FOR FACTORY SYNC -- %q", n)
		res := fac.WaitForCacheSync(f.stopChan)
		for k, v := range res {
			r[k] = v
			log.Debug().Msgf("  GVR resource %v -- %v", k, v)
		}
		log.Debug().Msgf("<<< DONE!")
	}
	return r
}

func (f *Factory) Init() {
	f.Start(f.stopChan)
}

func (f *Factory) Terminate() {
	if f.stopChan != nil {
		close(f.stopChan)
		f.stopChan = nil
	}
	for k := range f.factories {
		delete(f.factories, k)
	}
}

// Start initializes the informers until caller cancels the context.
func (f *Factory) Start(stopChan chan struct{}) {
	for ns, fac := range f.factories {
		log.Debug().Msgf("Starting factory in ns %q", ns)
		fac.Start(stopChan)
	}
}

// BOZO!! Check ns access for resource??
func (f *Factory) SetActive(ns string) {
	if !f.isClusterWide() {
		f.ensureFactory(ns)
	}
	f.activeNS = ns
}

func (f *Factory) isClusterWide() bool {
	_, ok := f.factories[render.AllNamespaces]
	return ok
}

func (f *Factory) preload(ns string) {
	f.ForResource(ns, "v1/pods")
	f.ForResource(render.AllNamespaces, "apiextensions.k8s.io/v1beta1/customresourcedefinitions")
}

func (f *Factory) FactoryFor(ns string) di.DynamicSharedInformerFactory {
	return f.factories[ns]
}

func (f *Factory) Preload(ns, gvr string) {
	_ = f.ForResource(ns, gvr)
}

func (f *Factory) ForResource(ns, gvr string) informers.GenericInformer {
	defer func(t time.Time) {
		log.Debug().Msgf("ForResource Elapsed %v", time.Since(t))
	}(time.Now())

	fact := f.ensureFactory(ns)
	log.Debug().Msgf("--- FORRESOURCE %q -- %q -- %#v", ns, gvr, toGVR(gvr))
	inf := fact.ForResource(toGVR(gvr))
	fact.Start(f.stopChan)

	return inf
}

func (f *Factory) ensureFactory(ns string) di.DynamicSharedInformerFactory {
	if f.isClusterWide() {
		ns = render.AllNamespaces
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
	// f.WaitForCacheSync()
	f.Dump()

	return f.factories[ns]
}

func (f *Factory) register(gvr, ns string, stopChan <-chan struct{}) error {
	log.Debug().Msgf("Registering GVR %q - %s", ns, gvr)
	f.factories[ns].ForResource(toGVR(gvr))
	f.factories[ns].Start(stopChan)

	return nil
}

func toGVR(s string) schema.GroupVersionResource {
	tokens := strings.Split(s, "/")
	if len(tokens) < 3 {
		tokens = append([]string{""}, tokens...)
	}

	return schema.GroupVersionResource{
		Group:    tokens[0],
		Version:  tokens[1],
		Resource: tokens[2],
	}
}

// Client return the factory connection.
func (f *Factory) Client() k8s.Connection {
	return f.client
}
