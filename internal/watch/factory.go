package watch

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/derailed/k9s/internal/k9s"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	di "k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/informers/internalinterfaces"
)

const defaultResync = 10 * time.Minute

// Factory tracks various resource informers.
type Factory struct {
	factories        map[string]di.DynamicSharedInformerFactory
	client           k9s.Connection
	stopChan         chan struct{}
	tweakListOptions internalinterfaces.TweakListOptionsFunc
	activeNS         string
}

// NewFactory returns a new informers factory.
func NewFactory(client k9s.Connection) *Factory {
	return &Factory{
		client:    client,
		stopChan:  make(chan struct{}),
		factories: make(map[string]di.DynamicSharedInformerFactory),
	}
}

func (f *Factory) Dump() {
	log.Debug().Msgf("----------- FACTORIES -------------")
	for ns := range f.factories {
		log.Debug().Msgf("Factory for NS %q", ns)
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

	return inf.Lister().ByNamespace(ns).List(sel)
}

func (f *Factory) Get(ns, gvr, name string, sel labels.Selector) (runtime.Object, error) {
	log.Debug().Msgf("<<<<<<<<<<<<<<<<< FACTORY GET %q", gvr)
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

	return inf.Lister().ByNamespace(ns).Get(name)
}

func (f *Factory) WaitForCacheSync() map[schema.GroupVersionResource]bool {
	r := make(map[schema.GroupVersionResource]bool)
	for n, fac := range f.factories {
		log.Debug().Msgf("Waiting for fac %q", n)
		res := fac.WaitForCacheSync(f.stopChan)
		log.Debug().Msgf("DONE!")
		for k, v := range res {
			r[k] = v
			log.Debug().Msgf("CACHE %v -- %v", k, v)
		}
	}
	return r
}

func (f *Factory) Init(ctx context.Context) {
	go func() {
		f.Start(f.stopChan)
		<-ctx.Done()
		f.Terminate()
	}()
}

func (f *Factory) Terminate() {
	if f.stopChan != nil {
		close(f.stopChan)
		f.stopChan = nil
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
	if !f.cluserWide() {
		f.ensureFactory(ns)
	}
	f.activeNS = ns
}

func (f *Factory) cluserWide() bool {
	_, ok := f.factories[""]
	return ok
}

func (f *Factory) ensureFactory(ns string) di.DynamicSharedInformerFactory {
	if f.cluserWide() {
		ns = ""
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
	f.WaitForCacheSync()
	f.Dump()

	return f.factories[ns]
}

func (f *Factory) preload(ns string) {
	f.ForResource(ns, "v1/pods")
	f.ForResource("", "apiextensions.k8s.io/v1beta1/customresourcedefinitions")
}

func (f *Factory) FactoryFor(ns string) di.DynamicSharedInformerFactory {
	return f.factories[ns]
}

func (f *Factory) ForResource(ns, gvr string) informers.GenericInformer {
	log.Debug().Msgf("Loading resource %q", gvr)
	fact := f.ensureFactory(ns)
	log.Debug().Msgf("--- FORRESOURCE %q -- %#v", ns, toGVR(gvr))
	inf := fact.ForResource(toGVR(gvr))
	fact.Start(f.stopChan)

	// f.WaitForCacheSync()
	// for i, k := range inf.Informer().GetStore().ListKeys() {
	// 	log.Debug().Msgf("%d -- %s", i, k)
	// }

	return inf
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
func (f *Factory) Client() k9s.Connection {
	return f.client
}

// func (f *Factory) ForResource(res schema.GroupVersionResource) informers.GenericInformer {
// 	log.Debug().Msgf("ForResource %v", res)
// 	switch res {
// 	case schema.GroupVersionResource{"metrics.k8s.io", "v1beta1", "pods"}:
// 		return &genericInformer{
// 			resource: res.GroupResource(),
// 			informer: f.MetricsV1Beta1("").PodMetricses().Informer(),
// 		}
// 	case schema.GroupVersionResource{"metrics.k8s.io", "v1beta1", "nodes"}:
// 		return &genericInformer{
// 			resource: res.GroupResource(),
// 			informer: f.MetricsV1Beta1("").NodeMetricses().Informer(),
// 		}
// 	default:
// 		return f.factories[""].ForResource(res)
// 	}
// }

// func (f *Factory) MetricsV1Beta1(ns string) v1beta1.Interface {
// 	return v1beta1.New(f.client, f, ns, f.tweakListOptions)
// }

// type genericInformer struct {
// 	informer cache.SharedIndexInformer
// 	resource schema.GroupResource
// }

// // Informer returns the SharedIndexInformer.
// func (f *genericInformer) Informer() cache.SharedIndexInformer {
// 	return f.informer
// }

// // Lister returns the GenericLister.
// func (f *genericInformer) Lister() cache.GenericLister {
// 	return cache.NewGenericLister(f.Informer().GetIndexer(), f.resource)
// }

// // InternalInformerFor returns the SharedIndexInformer for obj using an internal
// // client.
// func (f *Factory) InformerFor(obj runtime.Object, newFunc internalinterfaces.NewInformerFunc) cache.SharedIndexInformer {
// 	// var inf informers.GenericInformer

// 	kind := reflect.TypeOf(obj)
// 	log.Debug().Msgf("Informer for %v", kind)
// 	// switch kind {
// 	// case v1beta1.PodMetrics:
// 	// 	inf = f.ForResource("", toGVR("metrics.k8s.io/v1beta1/pods"))
// 	// 	if inf, ok := f.informers[kind]; ok {
// 	// 		return inf
// 	// 	}
// 	// case v1beta1.NodeMetrics:
// 	// 	inf = f.ForResource("", toGVR("metrics.k8s.io/v1beta1/nodes"))
// 	// 	if inf, ok := f.informers[kind]; ok {
// 	// 		return inf
// 	// 	}
// 	// default:
// 	// 	panic(fmt.Errorf("Unknown type %#v", t))
// 	// }
// 	// informerType :=
// 	// informer, exists := f.informers[informerType]
// 	// if exists {
// 	// 	return informer
// 	// }

// 	// resyncPeriod, exists := f.customResync[informerType]
// 	// if !exists {
// 	// 	resyncPeriod = f.defaultResync
// 	// }

// 	// informer = newFunc(f.client, resyncPeriod)
// 	// f.informers[kind] = informer

// 	// return informer
// 	return nil
// }
