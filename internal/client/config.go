// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

const (
	// DefaultCallTimeoutDuration is the default api server call timeout duration.
	DefaultCallTimeoutDuration time.Duration = 120 * time.Second

	// UsePersistentConfig caches client config to avoid reloads.
	UsePersistentConfig = true

	// InClusterContext names the synthesized context when running in a pod
	// with no kubeconfig, using the mounted service account credentials.
	InClusterContext = "in-cluster"

	defaultQPS   float32 = 50
	defaultBurst         = 100
)

// saDir is a var so tests can point at a fake service account mount.
var saDir = "/var/run/secrets/kubernetes.io/serviceaccount"

// Config tracks a kubernetes configuration.
type Config struct {
	flags  *genericclioptions.ConfigFlags
	mx     sync.RWMutex
	proxy  func(*http.Request) (*url.URL, error)
	icOnce sync.Once
	ic     bool
}

// NewConfig returns a new k8s config or an error if the flags are invalid.
func NewConfig(f *genericclioptions.ConfigFlags) *Config {
	return &Config{
		flags: f,
	}
}

// CallTimeout returns the call timeout if set or the default if not set.
func (c *Config) CallTimeout() time.Duration {
	if !isSet(c.flags.Timeout) {
		return DefaultCallTimeoutDuration
	}
	dur, err := time.ParseDuration(*c.flags.Timeout)
	if err != nil {
		return DefaultCallTimeoutDuration
	}

	return dur
}

func (c *Config) RESTConfig() (*restclient.Config, error) {
	cfg, err := c.clientConfig().ClientConfig()
	if err != nil {
		return nil, err
	}
	if c.proxy != nil {
		cfg.Proxy = c.proxy
	}
	if cfg.QPS == 0 {
		cfg.QPS = defaultQPS
		cfg.Burst = defaultBurst
	}

	return cfg, nil
}

// Flags returns configuration flags.
func (c *Config) Flags() *genericclioptions.ConfigFlags {
	return c.flags
}

func (c *Config) RawConfig() (api.Config, error) {
	return c.clientConfig().RawConfig()
}

func (c *Config) clientConfig() clientcmd.ClientConfig {
	if c.inCluster() {
		return clientcmd.NewNonInteractiveClientConfig(
			*inClusterKubeConfig(),
			InClusterContext,
			&clientcmd.ConfigOverrides{},
			&clientcmd.ClientConfigLoadingRules{},
		)
	}

	return c.flags.ToRawKubeConfigLoader()
}

// inCluster reports whether no kubeconfig nor api credentials were provided
// and the pod's service account is mounted, mirroring clientcmd's in-cluster
// fallback detection.
func (c *Config) inCluster() bool {
	c.icOnce.Do(func() {
		if c.flags == nil || isSet(c.flags.APIServer) || isSet(c.flags.BearerToken) {
			return
		}
		cfg, err := c.flags.ToRawKubeConfigLoader().RawConfig()
		if err != nil || len(cfg.Contexts) > 0 {
			return
		}
		fi, err := os.Stat(filepath.Join(saDir, "token"))
		c.ic = err == nil && !fi.IsDir() &&
			os.Getenv("KUBERNETES_SERVICE_HOST") != "" &&
			os.Getenv("KUBERNETES_SERVICE_PORT") != ""
	})

	return c.ic
}

// inClusterKubeConfig synthesizes a single-context kubeconfig from the pod's
// service account mount so the context machinery works without a kubeconfig.
func inClusterKubeConfig() *api.Config {
	ns := "default"
	if b, err := os.ReadFile(filepath.Join(saDir, "namespace")); err == nil {
		if s := strings.TrimSpace(string(b)); s != "" {
			ns = s
		}
	}
	cfg := api.NewConfig()
	cfg.Clusters[InClusterContext] = &api.Cluster{
		Server: "https://" + net.JoinHostPort(
			os.Getenv("KUBERNETES_SERVICE_HOST"),
			os.Getenv("KUBERNETES_SERVICE_PORT"),
		),
		CertificateAuthority: filepath.Join(saDir, "ca.crt"),
	}
	cfg.AuthInfos[InClusterContext] = &api.AuthInfo{
		TokenFile: filepath.Join(saDir, "token"),
	}
	cfg.Contexts[InClusterContext] = &api.Context{
		Cluster:   InClusterContext,
		AuthInfo:  InClusterContext,
		Namespace: ns,
	}
	cfg.CurrentContext = InClusterContext

	return cfg
}

func (*Config) reset() {}

// SwitchContext changes the kubeconfig context to a new cluster.
func (c *Config) SwitchContext(name string) error {
	ct, err := c.GetContext(name)
	if err != nil {
		return fmt.Errorf("context %q does not exist", name)
	}
	// !!BOZO!! Do you need to reset the flags?
	flags := genericclioptions.NewConfigFlags(UsePersistentConfig)
	flags.Context, flags.ClusterName = &name, &ct.Cluster
	flags.Namespace = c.flags.Namespace
	flags.Timeout = c.flags.Timeout
	flags.KubeConfig = c.flags.KubeConfig
	flags.Impersonate = c.flags.Impersonate
	flags.ImpersonateGroup = c.flags.ImpersonateGroup
	flags.ImpersonateUID = c.flags.ImpersonateUID
	flags.Insecure = c.flags.Insecure
	flags.BearerToken = c.flags.BearerToken

	c.flags = flags

	return nil
}

func (c *Config) Clone(ns string) (*genericclioptions.ConfigFlags, error) {
	flags := genericclioptions.NewConfigFlags(false)
	ct, err := c.CurrentContextName()
	if err != nil {
		return nil, err
	}
	cl, err := c.CurrentClusterName()
	if err != nil {
		return nil, err
	}
	flags.Context, flags.ClusterName = &ct, &cl
	flags.Namespace = &ns
	flags.Timeout = c.Flags().Timeout
	flags.KubeConfig = c.Flags().KubeConfig

	return flags, nil
}

// CurrentClusterName returns the currently active cluster name.
func (c *Config) CurrentClusterName() (string, error) {
	if isSet(c.flags.ClusterName) {
		return *c.flags.ClusterName, nil
	}
	cfg, err := c.RawConfig()
	if err != nil {
		return "", err
	}

	ct, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		return "", fmt.Errorf("invalid current context specified: %q", cfg.CurrentContext)
	}
	if isSet(c.flags.Context) {
		ct, ok = cfg.Contexts[*c.flags.Context]
		if !ok {
			return "", fmt.Errorf("current-cluster - invalid context specified: %q", *c.flags.Context)
		}
	}

	return ct.Cluster, nil
}

// CurrentContextName returns the currently active config context.
func (c *Config) CurrentContextName() (string, error) {
	if isSet(c.flags.Context) {
		return *c.flags.Context, nil
	}
	cfg, err := c.RawConfig()
	if err != nil {
		return "", fmt.Errorf("fail to load rawConfig: %w", err)
	}

	return cfg.CurrentContext, nil
}

func (c *Config) CurrentContextNamespace() (string, error) {
	name, err := c.CurrentContextName()
	if err != nil {
		return "", err
	}
	context, err := c.GetContext(name)
	if err != nil {
		return "", err
	}

	return context.Namespace, nil
}

// CurrentContext returns the current context configuration.
func (c *Config) CurrentContext() (*api.Context, error) {
	n, err := c.CurrentContextName()
	if err != nil {
		return nil, err
	}
	return c.GetContext(n)
}

// GetContext fetch a given context or error if it does not exist.
func (c *Config) GetContext(n string) (*api.Context, error) {
	cfg, err := c.RawConfig()
	if err != nil {
		return nil, err
	}
	if c, ok := cfg.Contexts[n]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("getcontext - invalid context specified: %q", n)
}

// SetProxy sets the proxy function.
func (c *Config) SetProxy(proxy func(*http.Request) (*url.URL, error)) {
	c.proxy = proxy
}

// Contexts fetch all available contexts.
func (c *Config) Contexts() (map[string]*api.Context, error) {
	cfg, err := c.RawConfig()
	if err != nil {
		return nil, err
	}

	return cfg.Contexts, nil
}

// DelContext remove a given context from the configuration.
func (c *Config) DelContext(n string) error {
	cfg, err := c.RawConfig()
	if err != nil {
		return err
	}
	delete(cfg.Contexts, n)

	acc, err := c.ConfigAccess()
	if err != nil {
		return err
	}

	return clientcmd.ModifyConfig(acc, cfg, true)
}

// RenameContext renames a context.
func (c *Config) RenameContext(oldCtx, newCtx string) error {
	cfg, err := c.RawConfig()
	if err != nil {
		return err
	}

	if _, ok := cfg.Contexts[newCtx]; ok {
		return fmt.Errorf("context with name %s already exists", newCtx)
	}
	cfg.Contexts[newCtx] = cfg.Contexts[oldCtx]
	delete(cfg.Contexts, oldCtx)
	acc, err := c.ConfigAccess()
	if err != nil {
		return err
	}
	if e := clientcmd.ModifyConfig(acc, cfg, true); e != nil {
		return e
	}
	current, err := c.CurrentContextName()
	if err != nil {
		return err
	}
	if current == oldCtx {
		return c.SwitchContext(newCtx)
	}

	return nil
}

// ContextNames fetch all available contexts.
func (c *Config) ContextNames() (map[string]struct{}, error) {
	cfg, err := c.RawConfig()
	if err != nil {
		return nil, err
	}
	cc := make(map[string]struct{}, len(cfg.Contexts))
	for n := range cfg.Contexts {
		cc[n] = struct{}{}
	}

	return cc, nil
}

// CurrentGroupNames retrieves the active group names.
func (c *Config) CurrentGroupNames() ([]string, error) {
	if areSet(c.flags.ImpersonateGroup) {
		return *c.flags.ImpersonateGroup, nil
	}

	return []string{}, errors.New("unable to locate current group")
}

// ImpersonateGroups retrieves the active groups if set on the CLI.
func (c *Config) ImpersonateGroups() (string, error) {
	if areSet(c.flags.ImpersonateGroup) {
		return strings.Join(*c.flags.ImpersonateGroup, ","), nil
	}

	return "", errors.New("no groups set")
}

// ImpersonateUser retrieves the active user name if set on the CLI.
func (c *Config) ImpersonateUser() (string, error) {
	if isSet(c.flags.Impersonate) {
		return *c.flags.Impersonate, nil
	}

	return "", errors.New("no user set")
}

// CurrentUserName retrieves the active user name.
func (c *Config) CurrentUserName() (string, error) {
	if isSet(c.flags.Impersonate) {
		return *c.flags.Impersonate, nil
	}

	if isSet(c.flags.AuthInfoName) {
		return *c.flags.AuthInfoName, nil
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "", err
	}

	current := cfg.CurrentContext
	if isSet(c.flags.Context) {
		current = *c.flags.Context
	}
	if ctx, ok := cfg.Contexts[current]; ok {
		return ctx.AuthInfo, nil
	}

	return "", errors.New("unable to locate current user")
}

// CurrentNamespaceName retrieves the active namespace.
func (c *Config) CurrentNamespaceName() (string, error) {
	ns, overridden, err := c.clientConfig().Namespace()
	if err != nil {
		return BlankNamespace, err
	}
	// Checks if ns is passed is in args.
	if overridden {
		return ns, nil
	}

	// Return ns set in context if any??
	return c.CurrentContextNamespace()
}

// ConfigAccess return the current kubeconfig api server access configuration.
func (c *Config) ConfigAccess() (clientcmd.ConfigAccess, error) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.clientConfig().ConfigAccess(), nil
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && *s != ""
}

func areSet(ss *[]string) bool {
	return ss != nil && len(*ss) != 0
}
