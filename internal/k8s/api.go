package k8s

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/client-go/discovery/cached/disk"

	"github.com/rs/zerolog/log"
	authorizationv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NA Not available
const NA = "n/a"

var supportedMetricsAPIVersions = []string{"v1beta1"}

type (
	// Collection of empty interfaces.
	Collection []interface{}

	// Cruder represent a crudable Kubernetes resource.
	Cruder interface {
		Get(ns string, name string) (interface{}, error)
		List(ns string) (Collection, error)
		Delete(ns string, name string) error
		SetFieldSelector(string)
		SetLabelSelector(string)
	}

	// Connection represents a Kubenetes apiserver connection.
	Connection interface {
		Config() *Config
		DialOrDie() kubernetes.Interface
		SwitchContextOrDie(ctx string)
		NSDialOrDie() dynamic.NamespaceableResourceInterface
		CachedDiscovery() (*disk.CachedDiscoveryClient, error)
		RestConfigOrDie() *restclient.Config
		MXDial() (*versioned.Clientset, error)
		DynDialOrDie() dynamic.Interface
		HasMetrics() bool
		IsNamespaced(n string) bool
		SupportsResource(group string) bool
		ValidNamespaces() ([]v1.Namespace, error)
		NodePods(node string) (*v1.PodList, error)
		SupportsRes(grp string, versions []string) (string, bool, error)
		ServerVersion() (*version.Info, error)
		FetchNodes() (*v1.NodeList, error)
		CurrentNamespaceName() (string, error)
		CheckNSAccess(ns string) error
		CheckListNSAccess() error
		CanIAccess(ns, rvg string, verbs []string) (bool, error)
	}

	clients struct {
		client          kubernetes.Interface
		dClient         dynamic.Interface
		nsClient        dynamic.NamespaceableResourceInterface
		mxsClient       *versioned.Clientset
		cachedDiscovery *disk.CachedDiscoveryClient
	}

	// APIClient represents a Kubernetes api client.
	APIClient struct {
		clients
		config          *Config
		useMetricServer bool
		mx              sync.Mutex
	}
)

// InitConnectionOrDie initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnectionOrDie(config *Config) *APIClient {
	conn := APIClient{config: config}
	conn.useMetricServer = conn.supportsMxServer()

	return &conn
}

// CheckListNSAccess check if current user can list namespaces.
func (a *APIClient) CheckListNSAccess() error {
	ns := NewNamespace(a)
	_, err := ns.List("", metav1.ListOptions{})
	return err
}

// CheckNSAccess asserts if user can access a namespace.
func (a *APIClient) CheckNSAccess(n string) error {
	ns := NewNamespace(a)
	if n == "" {
		_, err := ns.List(n, metav1.ListOptions{})
		return err
	}

	_, err := ns.Get("", n)
	return err
}

func makeSAR(ns, rvg string) *authorizationv1.SelfSubjectAccessReview {
	gvr, _ := schema.ParseResourceArg(strings.ToLower(rvg))
	if gvr == nil {
		panic(fmt.Errorf("Unable to get GVR from url %s", rvg))
	}
	log.Debug().Msgf("GVR for %s -- %#v", rvg, *gvr)
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace: ns,
				Group:     gvr.Group,
				Resource:  gvr.Resource,
			},
		},
	}
}

// CanIAccess checks if user has access to a certain resource.
func (a *APIClient) CanIAccess(ns, rvg string, verbs []string) (bool, error) {
	sar := makeSAR(ns, rvg)
	dial := a.DialOrDie().AuthorizationV1().SelfSubjectAccessReviews()
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := dial.Create(sar)
		if err != nil {
			log.Error().Err(err).Msgf("CanIAccess")
			return false, err
		}
		if !resp.Status.Allowed {
			return false, err
		}
	}
	return true, nil
}

// CurrentNamespaceName return namespace name set via either cli arg or cluster config.
func (a *APIClient) CurrentNamespaceName() (string, error) {
	return a.config.CurrentNamespaceName()
}

// ServerVersion returns the current server version info.
func (a *APIClient) ServerVersion() (*version.Info, error) {
	discovery, err := a.CachedDiscovery()
	if err != nil {
		return nil, err
	}

	return discovery.ServerVersion()
}

// FetchNodes returns all available nodes.
func (a *APIClient) FetchNodes() (*v1.NodeList, error) {
	return a.DialOrDie().CoreV1().Nodes().List(metav1.ListOptions{})
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	nn, err := a.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nn.Items, nil
}

// NodePods returns a collection of all available pods on a given node.
func (a *APIClient) NodePods(node string) (*v1.PodList, error) {
	const selFmt = "spec.nodeName=%s,status.phase!=%s,status.phase!=%s"
	fieldSelector, err := fields.ParseSelector(fmt.Sprintf(selFmt, node, v1.PodSucceeded, v1.PodFailed))
	if err != nil {
		return nil, err
	}

	return a.DialOrDie().CoreV1().Pods("").List(metav1.ListOptions{
		FieldSelector: fieldSelector.String(),
	})
}

// IsNamespaced check on server if given resource is namespaced
func (a *APIClient) IsNamespaced(res string) bool {
	discovery, err := a.CachedDiscovery()
	if err != nil {
		return false
	}

	list, _ := discovery.ServerPreferredResources()
	for _, l := range list {
		for _, r := range l.APIResources {
			if r.Name == res {
				return r.Namespaced
			}
		}
	}
	return false
}

// SupportsResource checks for resource supported version against the server.
func (a *APIClient) SupportsResource(group string) bool {
	discovery, err := a.CachedDiscovery()
	if err != nil {
		return false
	}

	list, err := discovery.ServerPreferredResources()
	if err != nil {
		log.Error().Err(err).Msg("Unable to dial api server")
		return false
	}
	for _, l := range list {
		log.Debug().Msgf(">>> Group %s", l.GroupVersion)
		if l.GroupVersion == group {
			return true
		}
	}
	return false
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() *Config {
	return a.config
}

// HasMetrics returns true if the cluster supports metrics.
func (a *APIClient) HasMetrics() bool {
	return a.useMetricServer
}

// DialOrDie returns a handle to api server or die.
func (a *APIClient) DialOrDie() kubernetes.Interface {
	if a.client != nil {
		return a.client
	}

	var err error
	if a.client, err = kubernetes.NewForConfig(a.RestConfigOrDie()); err != nil {
		log.Fatal().Msgf("Unable to connect to api server %v", err)
	}
	return a.client
}

// RestConfigOrDie returns a rest api client.
func (a *APIClient) RestConfigOrDie() *restclient.Config {
	cfg, err := a.config.RESTConfig()
	if err != nil {
		log.Panic().Msgf("Unable to connect to api server %v", err)
	}
	return cfg
}

func (a *APIClient) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.cachedDiscovery != nil {
		return a.cachedDiscovery, nil
	}

	rc := a.RestConfigOrDie()
	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(rc.Host))

	var err error
	a.cachedDiscovery, err = disk.NewCachedDiscoveryClientForConfig(rc, discCacheDir, httpCacheDir, 10*time.Minute)
	return a.cachedDiscovery, err
}

// DynDialOrDie returns a handle to a dynamic interface.
func (a *APIClient) DynDialOrDie() dynamic.Interface {
	if a.dClient != nil {
		return a.dClient
	}

	var err error
	if a.dClient, err = dynamic.NewForConfig(a.RestConfigOrDie()); err != nil {
		log.Panic().Err(err)
	}
	return a.dClient
}

// NSDialOrDie returns a handle to a namespaced resource.
func (a *APIClient) NSDialOrDie() dynamic.NamespaceableResourceInterface {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.nsClient != nil {
		return a.nsClient
	}
	a.nsClient = a.DynDialOrDie().Resource(schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1beta1",
		Resource: "customresourcedefinitions",
	})

	return a.nsClient
}

// MXDial returns a handle to the metrics server.
func (a *APIClient) MXDial() (*versioned.Clientset, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.mxsClient != nil {
		return a.mxsClient, nil
	}
	var err error
	if a.mxsClient, err = versioned.NewForConfig(a.RestConfigOrDie()); err != nil {
		log.Error().Err(err)
	}

	return a.mxsClient, err
}

// SwitchContextOrDie handles kubeconfig context switches.
func (a *APIClient) SwitchContextOrDie(ctx string) {
	currentCtx, err := a.config.CurrentContextName()
	if err != nil {
		panic(err)
	}

	if currentCtx != ctx {
		a.cachedDiscovery = nil
		a.reset()
		if err := a.config.SwitchContext(ctx); err != nil {
			panic(err)
		}
		a.useMetricServer = a.supportsMxServer()
	}
}

func (a *APIClient) reset() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.client, a.dClient, a.nsClient, a.mxsClient = nil, nil, nil, nil
}

func (a *APIClient) supportsMxServer() bool {
	discovery, err := a.CachedDiscovery()
	if err != nil {
		return false
	}

	apiGroups, err := discovery.ServerGroups()
	if err != nil {
		return false
	}

	for _, grp := range apiGroups.Groups {
		if grp.Name != metricsapi.GroupName {
			continue
		}
		if checkMetricsVersion(grp) {
			return true
		}
	}

	return false
}

func checkMetricsVersion(grp metav1.APIGroup) bool {
	for _, version := range grp.Versions {
		for _, supportedVersion := range supportedMetricsAPIVersions {
			if version.Version == supportedVersion {
				return true
			}
		}
	}

	return false
}

// SupportsRes checks latest supported version.
func (a *APIClient) SupportsRes(group string, versions []string) (string, bool, error) {
	discovery, err := a.CachedDiscovery()
	if err != nil {
		return "", false, err
	}

	apiGroups, err := discovery.ServerGroups()
	if err != nil {
		return "", false, err
	}

	for _, grp := range apiGroups.Groups {
		if grp.Name != group {
			continue
		}
		return grp.Versions[len(grp.Versions)-1].Version, true, nil
	}

	return "", false, nil
}
