package client

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	authorizationv1 "k8s.io/api/authorization/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	cacheSize   = 100
	cacheExpiry = 5 * time.Minute
	cacheMXKey  = "metrics"
)

var supportedMetricsAPIVersions = []string{"v1beta1"}

// Authorizer checks what a user can or cannot do to a resource.
type Authorizer interface {
	// CanI returns true if the user can use these actions for a given resource.
	CanI(ns, gvr string, verbs []string) (bool, error)
}

// APIClient represents a Kubernetes api client.
type APIClient struct {
	client       kubernetes.Interface
	dClient      dynamic.Interface
	nsClient     dynamic.NamespaceableResourceInterface
	mxsClient    *versioned.Clientset
	cachedClient *disk.CachedDiscoveryClient
	config       *Config
	mx           sync.Mutex
	cache        *cache.LRUExpireCache
}

// InitConnectionOrDie initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnectionOrDie(config *Config) *APIClient {
	a := APIClient{
		config: config,
		cache:  cache.NewLRUExpireCache(cacheSize),
	}
	a.HasMetrics()

	return &a
}

func makeSAR(ns, gvr string) *authorizationv1.SelfSubjectAccessReview {
	if ns == "-" {
		ns = ""
	}
	spec := NewGVR(gvr)
	res := spec.AsGVR()
	return &authorizationv1.SelfSubjectAccessReview{
		Spec: authorizationv1.SelfSubjectAccessReviewSpec{
			ResourceAttributes: &authorizationv1.ResourceAttributes{
				Namespace:   ns,
				Group:       res.Group,
				Resource:    res.Resource,
				Subresource: spec.SubResource(),
			},
		},
	}
}

func makeCacheKey(ns, gvr string, vv []string) string {
	return ns + ":" + gvr + "::" + strings.Join(vv, ",")
}

// CanI checks if user has access to a certain resource.
func (a *APIClient) CanI(ns, gvr string, verbs []string) (auth bool, err error) {
	if IsClusterWide(ns) {
		ns = AllNamespaces
	}
	key := makeCacheKey(ns, gvr, verbs)
	if v, ok := a.cache.Get(key); ok {
		if auth, ok = v.(bool); ok {
			return auth, nil
		}
	}
	dial, sar := a.DialOrDie().AuthorizationV1().SelfSubjectAccessReviews(), makeSAR(ns, gvr)
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := dial.Create(sar)
		if err != nil {
			log.Warn().Err(err).Msgf("  Dial Failed!")
			a.cache.Add(key, false, cacheExpiry)
			return auth, err
		}
		if !resp.Status.Allowed {
			a.cache.Add(key, false, cacheExpiry)
			return auth, fmt.Errorf("`%s access denied for user on %q:%s", v, ns, gvr)
		}
	}

	auth = true
	a.cache.Add(key, true, cacheExpiry)
	return
}

// CurrentNamespaceName return namespace name set via either cli arg or cluster config.
func (a *APIClient) CurrentNamespaceName() (string, error) {
	return a.config.CurrentNamespaceName()
}

// ServerVersion returns the current server version info.
func (a *APIClient) ServerVersion() (*version.Info, error) {
	return a.CachedDiscoveryOrDie().ServerVersion()
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	nn, err := a.DialOrDie().CoreV1().Namespaces().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	return nn.Items, nil
}

// IsNamespaced check on server if given resource is namespaced
func (a *APIClient) IsNamespaced(res string) bool {
	list, _ := a.CachedDiscoveryOrDie().ServerPreferredResources()
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
	list, err := a.CachedDiscoveryOrDie().ServerPreferredResources()
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
	v, ok := a.cache.Get(cacheMXKey)
	if ok {
		flag, k := v.(bool)
		return k && flag
	}

	var flag bool
	dial, err := a.MXDial()
	if err != nil {
		a.cache.Add(cacheMXKey, flag, cacheExpiry)
		return flag
	}

	if _, err := dial.MetricsV1beta1().NodeMetricses().List(metav1.ListOptions{Limit: 1}); err == nil {
		flag = true
	}
	a.cache.Add(cacheMXKey, flag, cacheExpiry)

	return flag
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

// CachedDiscoveryOrDie returns a cached discovery client.
func (a *APIClient) CachedDiscoveryOrDie() *disk.CachedDiscoveryClient {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.cachedClient != nil {
		return a.cachedClient
	}

	rc := a.RestConfigOrDie()
	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(rc.Host))

	var err error
	a.cachedClient, err = disk.NewCachedDiscoveryClientForConfig(rc, discCacheDir, httpCacheDir, 10*time.Minute)
	if err != nil {
		log.Panic().Msgf("Unable to connect to discovery client %v", err)
	}
	return a.cachedClient
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
		log.Fatal().Err(err).Msg("Fetching current context")
	}

	if currentCtx != ctx {
		a.cachedClient = nil
		a.reset()
		if err := a.config.SwitchContext(ctx); err != nil {
			log.Fatal().Err(err).Msg("Switching context")
		}
		_ = a.supportsMxServer()
	}
}

func (a *APIClient) reset() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.client, a.dClient, a.nsClient, a.mxsClient = nil, nil, nil, nil
}

func (a *APIClient) supportsMxServer() (supported bool) {
	defer func() {
		a.cache.Add(cacheMXKey, supported, cacheExpiry)
	}()

	apiGroups, err := a.CachedDiscoveryOrDie().ServerGroups()
	if err != nil {
		return
	}
	for _, grp := range apiGroups.Groups {
		if grp.Name != metricsapi.GroupName {
			continue
		}
		if checkMetricsVersion(grp) {
			supported = true
			return
		}
	}

	return
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
	apiGroups, err := a.CachedDiscoveryOrDie().ServerGroups()
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
