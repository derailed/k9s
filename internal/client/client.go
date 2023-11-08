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
	cmdutil "k8s.io/kubectl/pkg/cmd/util"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	cacheSize     = 100
	cacheExpiry   = 5 * time.Minute
	cacheMXAPIKey = "metricsAPI"
	serverVersion = "serverVersion"
)

var supportedMetricsAPIVersions = []string{"v1beta1"}

// Namespaces tracks a collection of namespace names.
type Namespaces map[string]struct{}

// APIClient represents a Kubernetes api client.
type APIClient struct {
	client, logClient kubernetes.Interface
	dClient           dynamic.Interface
	nsClient          dynamic.NamespaceableResourceInterface
	mxsClient         *versioned.Clientset
	cachedClient      *disk.CachedDiscoveryClient
	config            *Config
	mx                sync.Mutex
	cache             *cache.LRUExpireCache
	connOK            bool
}

// NewTestAPIClient for testing ONLY!!
func NewTestAPIClient() *APIClient {
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
		connOK: true,
	}
	err := a.supportsMetricsResources()
	if err != nil {
		log.Error().Err(err).Msgf("Fail to locate metrics-server")
	}
	if errors.Is(err, noMetricServerErr) || errors.Is(err, metricsUnsupportedErr) {
		return &a, nil
	}
	a.connOK = false

	return &a, err
}

// ConnectionOK returns connection status.
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
				Version:     res.Version,
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
	if ns, err := a.CurrentNamespaceName(); err == nil {
		return ns
	}

	return AllNamespaces
}

func (a *APIClient) clearCache() {
	for _, k := range a.cache.Keys() {
		a.cache.Remove(k)
	}
}

// CanI checks if user has access to a certain resource.
func (a *APIClient) CanI(ns, gvr string, verbs []string) (auth bool, err error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if !a.connOK {
		return false, errors.New("ACCESS -- No API server connection")
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

	ctx, cancel := context.WithTimeout(context.Background(), a.config.CallTimeout())
	defer cancel()
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := client.Create(ctx, sar, metav1.CreateOptions{})
		log.Trace().Msgf("[CAN] %s(%s) %v <<%v>>", gvr, verbs, resp, err)
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
	if v, ok := a.cache.Get(serverVersion); ok {
		if vi, ok := v.(*version.Info); ok {
			return vi, nil
		}
	}
	dial, err := a.CachedDiscovery()
	if err != nil {
		return nil, err
	}

	info, err := dial.ServerVersion()
	if err != nil {
		return nil, err
	}
	a.cache.Add(serverVersion, info, cacheExpiry)

	return info, nil
}

// ValidNamespaces returns all available namespaces.
func (a *APIClient) ValidNamespaces() ([]v1.Namespace, error) {
	if a == nil {
		return nil, fmt.Errorf("validNamespaces: no available client found")
	}

	if nn, ok := a.cache.Get("validNamespaces"); ok {
		if nss, ok := nn.([]v1.Namespace); ok {
			return nss, nil
		}
	}
	dial, err := a.Dial()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), a.config.CallTimeout())
	defer cancel()
	nn, err := dial.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	a.cache.Add("validNamespaces", nn.Items, cacheExpiry)

	return nn.Items, nil
}

// CheckConnectivity return true if api server is cool or false otherwise.
func (a *APIClient) CheckConnectivity() bool {
	a.mx.Lock()
	defer a.mx.Unlock()

	defer func() {
		if err := recover(); err != nil {
			a.connOK = false
		}
		if !a.connOK {
			a.clearCache()
		}
	}()

	// Need reload to pick up any kubeconfig changes.
	cfg, err := NewConfig(a.config.flags).RESTConfig()
	if err != nil {
		log.Error().Err(err).Msgf("restConfig load failed")
		a.connOK = false
		return a.connOK
	}
	cfg.Timeout = a.config.CallTimeout()
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to connect to api server")
		a.connOK = false
		return a.connOK
	}

	// Check connection
	if _, err := client.ServerVersion(); err == nil {
		if !a.connOK {
			a.reset()
		}
	} else {
		log.Error().Err(err).Msgf("can't connect to cluster")
		a.connOK = false
	}

	return a.connOK
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() *Config {
	return a.config
}

// HasMetrics checks if the cluster supports metrics.
func (a *APIClient) HasMetrics() bool {
	err := a.supportsMetricsResources()
	if err != nil {
		log.Debug().Msgf("Metrics server detect failed: %s", err)
	}
	return err == nil
}

// DialLogs returns a handle to api server for logs.
func (a *APIClient) DialLogs() (kubernetes.Interface, error) {
	if !a.connOK {
		return nil, errors.New("no connection to dial")
	}
	if a.logClient != nil {
		return a.logClient, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	cfg.Timeout = 0
	if a.logClient, err = kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	}

	return a.logClient, nil
}

// Dial returns a handle to api server or die.
func (a *APIClient) Dial() (kubernetes.Interface, error) {
	if !a.connOK {
		return nil, errors.New("no connection to dial")
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

// RestConfig returns a rest api client.
func (a *APIClient) RestConfig() (*restclient.Config, error) {
	return a.config.RESTConfig()
}

// CachedDiscovery returns a cached discovery client.
func (a *APIClient) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	a.mx.Lock()
	defer a.mx.Unlock()

	if !a.connOK {
		return nil, errors.New("no connection to cached dial")
	}

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
	if err := a.config.SwitchContext(name); err != nil {
		return err
	}
	a.mx.Lock()
	{
		a.reset()
		ResetMetrics()
	}
	a.mx.Unlock()

	if !a.CheckConnectivity() {
		return fmt.Errorf("unable to connect to context %q", name)
	}

	return nil
}

func (a *APIClient) reset() {
	a.config.reset()
	a.cache = cache.NewLRUExpireCache(cacheSize)
	a.client, a.dClient, a.nsClient, a.mxsClient = nil, nil, nil, nil
	a.cachedClient, a.logClient = nil, nil
	a.connOK = true
}

func (a *APIClient) checkCacheBool(key string) (state bool, ok bool) {
	v, found := a.cache.Get(key)
	if !found {
		return
	}
	state, ok = v.(bool)
	return
}

func (a *APIClient) supportsMetricsResources() error {
	supported, ok := a.checkCacheBool(cacheMXAPIKey)
	if ok {
		if supported {
			return nil
		}
		return noMetricServerErr
	}

	defer func() {
		a.cache.Add(cacheMXAPIKey, supported, cacheExpiry)
	}()

	cfg := cmdutil.NewMatchVersionFlags(a.config.flags)
	f := cmdutil.NewFactory(cfg)
	dial, err := f.ToDiscoveryClient()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to dial discovery API")
		return err
	}
	apiGroups, err := dial.ServerGroups()
	if err != nil {
		return err
	}
	for _, grp := range apiGroups.Groups {
		if grp.Name != metricsapi.GroupName {
			continue
		}
		if checkMetricsVersion(grp) {
			supported = true
			return nil
		}
	}

	return metricsUnsupportedErr
}

func checkMetricsVersion(grp metav1.APIGroup) bool {
	for _, v := range grp.Versions {
		for _, supportedVersion := range supportedMetricsAPIVersions {
			if v.Version == supportedVersion {
				return true
			}
		}
	}

	return false
}
