// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
	authorizationv1 "k8s.io/api/authorization/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/cache"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"
	metricsapi "k8s.io/metrics/pkg/apis/metrics"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

const (
	cacheSize     = 100
	cacheExpiry   = 5 * time.Minute
	cacheMXAPIKey = "metricsAPI"
	serverVersion = "serverVersion"
	cacheNSKey    = "validNamespaces"
)

var supportedMetricsAPIVersions = []string{"v1beta1"}

// NamespaceNames tracks a collection of namespace names.
type NamespaceNames map[string]struct{}

// APIClient represents a Kubernetes api client.
type APIClient struct {
	client, logClient kubernetes.Interface
	dClient           dynamic.Interface
	nsClient          dynamic.NamespaceableResourceInterface
	mxsClient         *versioned.Clientset
	cachedClient      *disk.CachedDiscoveryClient
	config            *Config
	mx                sync.RWMutex
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
	if err == nil || errors.Is(err, noMetricServerErr) || errors.Is(err, metricsUnsupportedErr) {
		return &a, nil
	}
	a.connOK = false

	return &a, err
}

// ConnectionOK returns connection status.
func (a *APIClient) ConnectionOK() bool {
	return a.connOK
}

func makeSAR(ns, gvr, name string) *authorizationv1.SelfSubjectAccessReview {
	if ns == ClusterScope {
		ns = BlankNamespace
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
				Name:        name,
			},
		},
	}
}

func makeCacheKey(ns, gvr, n string, vv []string) string {
	return ns + ":" + gvr + ":" + n + "::" + strings.Join(vv, ",")
}

// ActiveContext returns the current context name.
func (a *APIClient) ActiveContext() string {
	c, err := a.config.CurrentContextName()
	if err != nil {
		log.Error().Msgf("Unable to located active cluster")
		return ""
	}
	return c
}

// IsActiveNamespace returns true if namespaces matches.
func (a *APIClient) IsActiveNamespace(ns string) bool {
	if a.ActiveNamespace() == BlankNamespace {
		return true
	}

	return a.ActiveNamespace() == ns
}

// ActiveNamespace returns the current namespace.
func (a *APIClient) ActiveNamespace() string {
	if ns, err := a.CurrentNamespaceName(); err == nil {
		return ns
	}

	return BlankNamespace
}

func (a *APIClient) clearCache() {
	for _, k := range a.cache.Keys() {
		a.cache.Remove(k)
	}
}

// CanI checks if user has access to a certain resource.
func (a *APIClient) CanI(ns, gvr, name string, verbs []string) (auth bool, err error) {
	if !a.getConnOK() {
		return false, errors.New("ACCESS -- No API server connection")
	}
	if IsClusterWide(ns) {
		ns = BlankNamespace
	}
	key := makeCacheKey(ns, gvr, name, verbs)
	if v, ok := a.cache.Get(key); ok {
		if auth, ok = v.(bool); ok {
			return auth, nil
		}
	}

	dial, err := a.Dial()
	if err != nil {
		return false, err
	}
	client, sar := dial.AuthorizationV1().SelfSubjectAccessReviews(), makeSAR(ns, gvr, name)

	ctx, cancel := context.WithTimeout(context.Background(), a.config.CallTimeout())
	defer cancel()
	for _, v := range verbs {
		sar.Spec.ResourceAttributes.Verb = v
		resp, err := client.Create(ctx, sar, metav1.CreateOptions{})
		log.Trace().Msgf("[CAN] %s(%q/%q) <%v>", gvr, ns, name, verbs)
		if resp != nil {
			log.Trace().Msgf("  Spec: %#v", resp.Spec)
			log.Trace().Msgf("  Auth: %t [%q]", resp.Status.Allowed, resp.Status.Reason)
		}
		log.Trace().Msgf("  <<%v>>", err)
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

func (a *APIClient) IsValidNamespace(ns string) bool {
	ok, err := a.isValidNamespace(ns)
	if err != nil {
		log.Warn().Err(err).Msgf("namespace validation failed for: %q", ns)
	}

	return ok
}

func (a *APIClient) isValidNamespace(n string) (bool, error) {
	if IsClusterWide(n) || n == NotNamespaced {
		return true, nil
	}
	nn, err := a.ValidNamespaceNames()
	if err != nil {
		return false, err
	}
	_, ok := nn[n]

	return ok, nil
}

// ValidNamespaceNames returns all available namespaces.
func (a *APIClient) ValidNamespaceNames() (NamespaceNames, error) {
	if a == nil {
		return nil, fmt.Errorf("validNamespaces: no available client found")
	}

	if nn, ok := a.cache.Get(cacheNSKey); ok {
		if nss, ok := nn.(NamespaceNames); ok {
			return nss, nil
		}
	}

	ok, err := a.CanI(ClusterScope, "v1/namespaces", "", ListAccess)
	if !ok || err != nil {
		return nil, fmt.Errorf("user not authorized to list all namespaces")
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
	nns := make(NamespaceNames, len(nn.Items))
	for _, n := range nn.Items {
		nns[n.Name] = struct{}{}
	}
	a.cache.Add(cacheNSKey, nns, cacheExpiry)

	return nns, nil
}

// CheckConnectivity return true if api server is cool or false otherwise.
func (a *APIClient) CheckConnectivity() bool {
	defer func() {
		if err := recover(); err != nil {
			a.setConnOK(false)
		}
		if !a.getConnOK() {
			a.clearCache()
		}
	}()

	cfg, err := a.config.RESTConfig()

	if err != nil {
		log.Error().Err(err).Msgf("restConfig load failed")
		a.connOK = false
		return a.connOK
	}
	cfg.Timeout = a.config.CallTimeout()
	client, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		log.Error().Err(err).Msgf("Unable to connect to api server")
		a.setConnOK(false)
		return a.getConnOK()
	}

	// Check connection
	if _, err := client.ServerVersion(); err == nil {
		if !a.getConnOK() {
			a.reset()
		}
	} else {
		log.Error().Err(err).Msgf("can't connect to cluster")
		a.setConnOK(false)
	}

	return a.getConnOK()
}

// Config return a kubernetes configuration.
func (a *APIClient) Config() *Config {
	return a.config
}

// HasMetrics checks if the cluster supports metrics.
func (a *APIClient) HasMetrics() bool {
	return a.supportsMetricsResources() == nil
}

func (a *APIClient) getMxsClient() *versioned.Clientset {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.mxsClient
}

func (a *APIClient) setMxsClient(c *versioned.Clientset) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.mxsClient = c
}

func (a *APIClient) getCachedClient() *disk.CachedDiscoveryClient {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.cachedClient
}

func (a *APIClient) setCachedClient(c *disk.CachedDiscoveryClient) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.cachedClient = c
}

func (a *APIClient) getDClient() dynamic.Interface {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.dClient
}

func (a *APIClient) setDClient(c dynamic.Interface) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.dClient = c
}

func (a *APIClient) getConnOK() bool {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.connOK
}

func (a *APIClient) setConnOK(b bool) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.connOK = b
}

func (a *APIClient) setLogClient(k kubernetes.Interface) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.logClient = k
}

func (a *APIClient) getLogClient() kubernetes.Interface {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.logClient
}

func (a *APIClient) setClient(k kubernetes.Interface) {
	a.mx.Lock()
	defer a.mx.Unlock()

	a.client = k
}

func (a *APIClient) getClient() kubernetes.Interface {
	a.mx.RLock()
	defer a.mx.RUnlock()

	return a.client
}

// DialLogs returns a handle to api server for logs.
func (a *APIClient) DialLogs() (kubernetes.Interface, error) {
	if !a.getConnOK() {
		return nil, errors.New("dialLogs - no connection to dial")
	}
	if clt := a.getLogClient(); clt != nil {
		return clt, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	cfg.Timeout = 0
	c, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	a.setLogClient(c)

	return a.getLogClient(), nil
}

// Dial returns a handle to api server or die.
func (a *APIClient) Dial() (kubernetes.Interface, error) {
	if !a.getConnOK() {
		return nil, errors.New("no connection to dial")
	}
	if c := a.getClient(); c != nil {
		return c, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	if c, err := kubernetes.NewForConfig(cfg); err != nil {
		return nil, err
	} else {
		a.setClient(c)
	}

	return a.getClient(), nil
}

// RestConfig returns a rest api client.
func (a *APIClient) RestConfig() (*restclient.Config, error) {
	return a.config.RESTConfig()
}

// CachedDiscovery returns a cached discovery client.
func (a *APIClient) CachedDiscovery() (*disk.CachedDiscoveryClient, error) {
	if !a.getConnOK() {
		return nil, errors.New("no connection to cached dial")
	}

	if c := a.getCachedClient(); c != nil {
		return c, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}

	baseCacheDir := os.Getenv("KUBECACHEDIR")
	if baseCacheDir == "" {
		baseCacheDir = filepath.Join(mustHomeDir(), ".kube", "cache")
	}

	httpCacheDir := filepath.Join(baseCacheDir, "http")
	discCacheDir := filepath.Join(baseCacheDir, "discovery", toHostDir(cfg.Host))

	c, err := disk.NewCachedDiscoveryClientForConfig(cfg, discCacheDir, httpCacheDir, cacheExpiry)
	if err != nil {
		return nil, err
	}
	a.setCachedClient(c)

	return a.getCachedClient(), nil
}

// DynDial returns a handle to a dynamic interface.
func (a *APIClient) DynDial() (dynamic.Interface, error) {
	if c := a.getDClient(); c != nil {
		return c, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}
	c, err := dynamic.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	a.setDClient(c)

	return a.getDClient(), nil
}

// MXDial returns a handle to the metrics server.
func (a *APIClient) MXDial() (*versioned.Clientset, error) {
	if c := a.getMxsClient(); c != nil {
		return c, nil
	}

	cfg, err := a.RestConfig()
	if err != nil {
		return nil, err
	}

	c, err := versioned.NewForConfig(cfg)
	if err != nil {
		return nil, err
	}
	a.setMxsClient(c)

	return a.getMxsClient(), err
}

func (a *APIClient) invalidateCache() error {
	dial, err := a.CachedDiscovery()
	if err != nil {
		return err
	}
	dial.Invalidate()

	return nil
}

// SwitchContext handles kubeconfig context switches.
func (a *APIClient) SwitchContext(name string) error {
	log.Debug().Msgf("Switching context %q", name)
	if err := a.config.SwitchContext(name); err != nil {
		return err
	}
	a.reset()
	ResetMetrics()

	// Need reload to pick up any kubeconfig changes.
	a.config = NewConfig(a.config.flags)

	return a.invalidateCache()
}

func (a *APIClient) reset() {
	a.config.reset()
	a.cache = cache.NewLRUExpireCache(cacheSize)
	a.nsClient = nil

	a.setDClient(nil)
	a.setMxsClient(nil)
	a.setCachedClient(nil)
	a.setClient(nil)
	a.setLogClient(nil)
	a.setConnOK(true)
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

	dial, err := a.Dial()
	if err != nil {
		log.Warn().Err(err).Msgf("Unable to dial discovery API")
		return err
	}
	apiGroups, err := dial.Discovery().ServerGroups()
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
