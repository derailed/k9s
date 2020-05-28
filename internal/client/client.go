package client

import (
	"context"
	"errors"
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
	cacheSize        = 100
	cacheExpiry      = 5 * time.Minute
	cacheMXKey       = "metrics"
	cacheMXAPIKey    = "metricsAPI"
	checkConnTimeout = 5 * time.Second

	// CallTimeout represents default api call timeout.
	CallTimeout = 5 * time.Second
)

var supportedMetricsAPIVersions = []string{"v1beta1"}

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
	connOK       bool
}

// NewTestClient for testing ONLY!!
func NewTestClient() *APIClient {
	return &APIClient{
		config: NewConfig(nil),
		cache:  cache.NewLRUExpireCache(cacheSize),
	}
}

// InitConnection initialize connection from command line args.
// Checks for connectivity with the api server.
func InitConnection(config *Config) (*APIClient, error) {
	a := APIClient{
		config: config,
		cache:  cache.NewLRUExpireCache(cacheSize),
	}
	_, err := a.supportsMetricsResources()
	if err == nil {
		a.connOK = true
	}

	return &a, err
}

func (a *APIClient) ConnectionOK() bool {
	return a.connOK
}

func makeSAR(ns, gvr string) *authorizationv1.SelfSubjectAccessReview {
	if ns == ClusterScope {
		ns = AllNamespaces
	}
	spec := NewGVR(gvr)
	res := spec.GVR()
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

// ActiveCluster returns the current cluster name.
func (a *APIClient) ActiveCluster() string {
	c, err := a.config.CurrentClusterName()
	if err != nil {
		log.Error().Msgf("Unable to located active cluster")
		return ""
	}
	return c
}

// IsActiveNamespace returns true if namespaces matches.
func (a *APIClient) IsActiveNamespace(ns string) bool {
	if a.ActiveNamespace() == AllNamespaces {
		return true
	}
	return a.ActiveNamespace() == ns
}

// ActiveNamespace returns the current namespace.
func (a *APIClient) ActiveNamespace() string {
	ns, err := a.CurrentNamespaceName()
	if err != nil {
		return AllNamespaces
	}
	return ns
}

func (a *APIClient) clearCache() {
	for _, k := range a.cache.Keys() {
		a.cache.Remove(k)
	}
}

// CanI checks if user has access to a certain resource.
func (a *APIClient) CanI(ns, gvr string, verbs []string) (auth bool, err error) {
	log.Debug().Msgf("Check Access %q::%q", ns, gvr)
	if !a.connOK {
		return false, errors.New("no API server connection")
	}
	if IsClusterWide(ns) {
		ns = AllNamespaces
	}
	key := makeCacheKey(ns, gvr, verbs)
	if v, ok := a.cache.Get(key); ok {
		if auth, ok = v.(bool); ok {
			return auth, nil
		}
	}

	dial, err := a.Dial()
	if err != nil {
		return false, err
	}
	client, sar := dial.AuthorizationV1().SelfSubjectAccessReviews(), makeSAR(ns, gvr)

	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	defer cancel()
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := client.Create(ctx, sar, metav1.CreateOptions{})
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
	dial, err := a.CachedDiscovery()
	if err != nil {
		return nil, err
	}
	return dial.ServerVersion()
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	if nn, ok := a.cache.Get("validNamespaces"); ok {
		if nss, ok := nn.([]v1.Namespace); ok {
			return nss, nil
		}
	}
	dial, err := a.Dial()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	defer cancel()
	nn, err := dial.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	a.cache.Add("validNamespaces", nn.Items, cacheExpiry)

	return nn.Items, nil
}

// CheckConnectivity return true if api server is cool or false otherwise.
func (a *APIClient) CheckConnectivity() (status bool) {
	defer func() {
		if err := recover(); err != nil {
			status = false
		}
		if !status {
			a.clearCache()
		}
		a.connOK = status
	}()

	// Need to reload to pickup any kubeconfig changes.
	config := NewConfig(a.config.flags)
	cfg, err := config.RESTConfig()
	if err != nil {
		return false
	}
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to connect to api server")
		return
	}
	log.Debug().Msgf("Checking APIServer on %#v", cfg.Host)

	// Check connection
	if _, err := client.ServerVersion(); err == nil {
		if !a.connOK {
			log.Debug().Msgf("RESETING CON!!")
			a.reset()
		}
		status = true
	} else {
		log.Error().Err(err).Msgf("K9s can't connect to cluster")
	}

	return
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() *Config {
	return a.config
}

// HasMetrics returns true if the cluster supports metrics.
func (a *APIClient) HasMetrics() bool {
	ok, err := a.supportsMetricsResources()
	if !ok || err != nil {
		return false
	}
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
	ctx, cancel := context.WithTimeout(context.Background(), CallTimeout)
	defer cancel()
	if _, err := dial.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{Limit: 1}); err == nil {
		flag = true
	} else {
		log.Error().Err(err).Msgf("List metrics failed")
	}
	a.cache.Add(cacheMXKey, flag, cacheExpiry)

	return flag
}

// Dial returns a handle to api server or die.
func (a *APIClient) Dial() (kubernetes.Interface, error) {
	if !a.connOK {
		return nil, errors.New("No connection to dial")
	}
	if a.client != nil {
		return a.client, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if a.client, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}

	return a.client, nil
}

// RestConfigOrDie returns a rest api client.
func (a *APIClient) RestConfig() (*restclient.Config, error) {
	cfg, err := a.config.RESTConfig()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

// CachedDiscovery returns a cached discovery client.
func (a *APIClient) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	if !a.connOK {
		return nil, errors.New("No connection to cached dial")
	}
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.cachedClient != nil {
		return a.cachedClient, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(cfg.Host))

	a.cachedClient, err = disk.NewCachedDiscoveryClientForConfig(cfg, discCacheDir, httpCacheDir, cacheExpiry)
	if err != nil {
		return nil, err
	}
	return a.cachedClient, nil
}

// DynDial returns a handle to a dynamic interface.
func (a *APIClient) DynDial() (dynamic.Interface, error) {
	if a.dClient != nil {
		return a.dClient, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if a.dClient, err = dynamic.NewForConfig(cfg); err != nil {
		log.Panic().Err(err)
	}

	return a.dClient, nil
}

// MXDial returns a handle to the metrics server.
func (a *APIClient) MXDial() (*versioned.Clientset, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if a.mxsClient != nil {
		return a.mxsClient, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}

	if a.mxsClient, err = versioned.NewForConfig(cfg); err != nil {
		log.Error().Err(err)
	}

	return a.mxsClient, err
}

// SwitchContext handles kubeconfig context switches.
func (a *APIClient) SwitchContext(name string) error {
	log.Debug().Msgf("Switching context %q", name)
	currentCtx, err := a.config.CurrentContextName()
	if err != nil {
		return err
	}
	if currentCtx == name {
		return nil
	}

	if e := a.config.SwitchContext(name); e != nil {
		return e
	}
	a.clearCache()
	a.reset()
	a.connOK = true
	_, err = a.supportsMetricsResources()
	if err != nil {
		a.connOK = false
		return err
	}
	ResetMetrics()

	return nil
}

func (a *APIClient) reset() {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.config.reset()
	a.cache = cache.NewLRUExpireCache(cacheSize)
	a.client, a.dClient, a.nsClient, a.mxsClient = nil, nil, nil, nil
	a.cachedClient = nil
}

func (a *APIClient) supportsMetricsResources() (supported bool, err error) {
	defer func() {
		a.cache.Add(cacheMXAPIKey, supported, cacheExpiry)
	}()

	if v, ok := a.cache.Get(cacheMXAPIKey); ok {
		flag, k := v.(bool)
		supported = k && flag
		return
	}
	if a.config == nil || a.config.flags == nil {
		return
	}

	dial, err := a.CachedDiscovery()
	if err != nil {
		return false, err
	}
	apiGroups, err := dial.ServerGroups()
	if err != nil {
		log.Debug().Msgf("Unable to access servergroups %#v", err)
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
