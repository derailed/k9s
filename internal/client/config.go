// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
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

	defaultQPS   float32 = 50
	defaultBurst         = 100
)

// Config tracks a kubernetes configuration.
type Config struct {
	// flags back the active connection and stay pinned to the active context.
	flags *genericclioptions.ConfigFlags
	// latest reflects the most recent kubeconfig content.
	latest *genericclioptions.ConfigFlags
	mx     sync.RWMutex
	proxy  func(*http.Request) (*url.URL, error)

	// Kubeconfig reload tracking.
	primed   bool
	stamps   []fileStamp
	active   api.Config
	revision uint64
	detached string
}

// NewConfig returns a new k8s config or an error if the flags are invalid.
func NewConfig(f *genericclioptions.ConfigFlags) *Config {
	return &Config{
		flags:  f,
		latest: f,
	}
}

// CallTimeout returns the call timeout if set or the default if not set.
func (c *Config) CallTimeout() time.Duration {
	f := c.Flags()
	if !isSet(f.Timeout) {
		return DefaultCallTimeoutDuration
	}
	dur, err := time.ParseDuration(*f.Timeout)
	if err != nil {
		return DefaultCallTimeoutDuration
	}

	return dur
}

func (c *Config) RESTConfig() (*restclient.Config, error) {
	_, loader := c.activeConfig()
	cfg, err := loader.ClientConfig()
	if err != nil {
		return nil, err
	}
	c.mx.RLock()
	proxy := c.proxy
	c.mx.RUnlock()
	if proxy != nil {
		cfg.Proxy = proxy
	}
	if cfg.QPS == 0 {
		cfg.QPS = defaultQPS
		cfg.Burst = defaultBurst
	}

	return cfg, nil
}

// Flags returns the active connection configuration flags.
func (c *Config) Flags() *genericclioptions.ConfigFlags {
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.flags
}

// RawConfig returns the latest kubeconfig content.
func (c *Config) RawConfig() (api.Config, error) {
	c.refresh()
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.latest.ToRawKubeConfigLoader().RawConfig()
}

// activeConfig returns the flags and loader backing the active connection.
func (c *Config) activeConfig() (*genericclioptions.ConfigFlags, clientcmd.ClientConfig) {
	c.refresh()
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.flags, c.flags.ToRawKubeConfigLoader()
}

func (*Config) reset() {}

// SwitchContext changes the kubeconfig context to a new cluster.
func (c *Config) SwitchContext(name string) error {
	ct, err := c.GetContext(name)
	if err != nil {
		return fmt.Errorf("context %q does not exist", name)
	}
	c.mx.Lock()
	defer c.mx.Unlock()

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

	c.flags, c.latest = flags, flags
	c.primed, c.detached = false, ""

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
	f := c.Flags()
	flags.Timeout = f.Timeout
	flags.KubeConfig = f.KubeConfig

	return flags, nil
}

// CurrentClusterName returns the currently active cluster name.
func (c *Config) CurrentClusterName() (string, error) {
	f, loader := c.activeConfig()
	if isSet(f.ClusterName) {
		return *f.ClusterName, nil
	}
	cfg, err := loader.RawConfig()
	if err != nil {
		return "", err
	}

	ct, ok := cfg.Contexts[cfg.CurrentContext]
	if !ok {
		return "", fmt.Errorf("invalid current context specified: %q", cfg.CurrentContext)
	}
	if isSet(f.Context) {
		ct, ok = cfg.Contexts[*f.Context]
		if !ok {
			return "", fmt.Errorf("current-cluster - invalid context specified: %q", *f.Context)
		}
	}

	return ct.Cluster, nil
}

// CurrentContextName returns the currently active config context.
func (c *Config) CurrentContextName() (string, error) {
	f, loader := c.activeConfig()
	if isSet(f.Context) {
		return *f.Context, nil
	}
	cfg, err := loader.RawConfig()
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
	if ct, ok := c.detachedContext(n); ok {
		return ct, nil
	}
	if c, ok := cfg.Contexts[n]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("getcontext - invalid context specified: %q", n)
}

// detachedContext returns the active context while the kubeconfig on disk no
// longer backs it, so K9s keeps running with the settings it loaded.
func (c *Config) detachedContext(n string) (*api.Context, bool) {
	c.mx.RLock()
	defer c.mx.RUnlock()

	if c.detached == "" {
		return nil, false
	}
	name := c.active.CurrentContext
	if isSet(c.flags.Context) {
		name = *c.flags.Context
	}
	if n != name {
		return nil, false
	}
	ct, ok := c.active.Contexts[n]

	return ct, ok
}

// SetProxy sets the proxy function.
func (c *Config) SetProxy(proxy func(*http.Request) (*url.URL, error)) {
	c.mx.Lock()
	defer c.mx.Unlock()

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
	if f := c.Flags(); areSet(f.ImpersonateGroup) {
		return *f.ImpersonateGroup, nil
	}

	return []string{}, errors.New("unable to locate current group")
}

// ImpersonateGroups retrieves the active groups if set on the CLI.
func (c *Config) ImpersonateGroups() (string, error) {
	if f := c.Flags(); areSet(f.ImpersonateGroup) {
		return strings.Join(*f.ImpersonateGroup, ","), nil
	}

	return "", errors.New("no groups set")
}

// ImpersonateUser retrieves the active user name if set on the CLI.
func (c *Config) ImpersonateUser() (string, error) {
	if f := c.Flags(); isSet(f.Impersonate) {
		return *f.Impersonate, nil
	}

	return "", errors.New("no user set")
}

// CurrentUserName retrieves the active user name.
func (c *Config) CurrentUserName() (string, error) {
	f, loader := c.activeConfig()
	if isSet(f.Impersonate) {
		return *f.Impersonate, nil
	}

	if isSet(f.AuthInfoName) {
		return *f.AuthInfoName, nil
	}

	cfg, err := loader.RawConfig()
	if err != nil {
		return "", err
	}

	current := cfg.CurrentContext
	if isSet(f.Context) {
		current = *f.Context
	}
	if ctx, ok := cfg.Contexts[current]; ok {
		return ctx.AuthInfo, nil
	}

	return "", errors.New("unable to locate current user")
}

// CurrentNamespaceName retrieves the active namespace.
func (c *Config) CurrentNamespaceName() (string, error) {
	_, loader := c.activeConfig()
	ns, overridden, err := loader.Namespace()
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
	c.refresh()
	c.mx.RLock()
	defer c.mx.RUnlock()

	return c.latest.ToRawKubeConfigLoader().ConfigAccess(), nil
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && *s != ""
}

func areSet(ss *[]string) bool {
	return ss != nil && len(*ss) != 0
}
