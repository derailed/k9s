package client

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const (
	defaultCallTimeoutDuration time.Duration = 5 * time.Second
)

// Config tracks a kubernetes configuration.
type Config struct {
	flags      *genericclioptions.ConfigFlags
	clientCfg  clientcmd.ClientConfig
	rawCfg     *clientcmdapi.Config
	mutex      *sync.RWMutex
	OverrideNS bool
}

// NewConfig returns a new k8s config or an error if the flags are invalid.
func NewConfig(f *genericclioptions.ConfigFlags) *Config {
	return &Config{
		flags: f,
		// pathOptions: clientcmd.NewDefaultPathOptions(),
		mutex: &sync.RWMutex{},
	}
}

// CallTimeout returns the call timeout if set or the default if not set.
func (c *Config) CallTimeout() time.Duration {
	if !isSet(c.flags.Timeout) {
		return defaultCallTimeoutDuration
	}
	dur, err := time.ParseDuration(*c.flags.Timeout)
	if err != nil {
		return defaultCallTimeoutDuration
	}

	return dur
}

func (c *Config) RESTConfig() (*restclient.Config, error) {
	return c.clientConfig().ClientConfig()
}

// Flags returns configuration flags.
func (c *Config) Flags() *genericclioptions.ConfigFlags {
	return c.flags
}

func (c *Config) rawConfig() (*clientcmdapi.Config, error) {
	if c.rawCfg != nil {
		return c.rawCfg, nil
	}

	cfg, err := c.clientConfig().RawConfig()
	if err != nil {
		return nil, err
	}
	c.rawCfg = &cfg

	return c.rawCfg, nil
}

func (c *Config) clientConfig() clientcmd.ClientConfig {
	if c.clientCfg != nil {
		return c.clientCfg
	}
	c.clientCfg = c.flags.ToRawKubeConfigLoader()

	return c.clientCfg
}

func (c *Config) reset() {
	c.clientCfg = nil
}

// SwitchContext changes the kubeconfig context to a new cluster.
func (c *Config) SwitchContext(name string) error {
	if n, err := c.CurrentContextName(); err == nil && n == name {
		return nil
	}
	context, err := c.GetContext(name)
	if err != nil {
		return fmt.Errorf("context %s does not exist", name)
	}
	c.flags.Context = &name
	c.flags.ClusterName = &(context.Cluster)

	return nil
}

func (c *Config) RawConfig() *clientcmdapi.Config {
	return c.rawCfg
}

// CurrentContextName returns the currently active config context.
func (c *Config) CurrentContextName() (string, error) {
	if isSet(c.flags.Context) {
		return *c.flags.Context, nil
	}
	cfg, err := c.rawConfig()
	if err != nil {
		return "", err
	}

	return cfg.CurrentContext, nil
}

// GetContext fetch a given context or error if it does not exists.
func (c *Config) GetContext(n string) (*clientcmdapi.Context, error) {
	cfg, err := c.rawConfig()
	if err != nil {
		return nil, err
	}
	if c, ok := cfg.Contexts[n]; ok {
		return c, nil
	}

	return nil, fmt.Errorf("invalid context `%s specified", n)
}

// Contexts fetch all available contexts.
func (c *Config) Contexts() (map[string]*clientcmdapi.Context, error) {
	cfg, err := c.rawConfig()
	if err != nil {
		return nil, err
	}

	return cfg.Contexts, nil
}

// DelContext remove a given context from the configuration.
func (c *Config) DelContext(n string) error {
	cfg, err := c.rawConfig()
	if err != nil {
		return err
	}
	delete(cfg.Contexts, n)

	acc, err := c.ConfigAccess()
	if err != nil {
		return err
	}

	return clientcmd.ModifyConfig(acc, *cfg, true)
}

// ContextNames fetch all available contexts.
func (c *Config) ContextNames() ([]string, error) {
	cfg, err := c.rawConfig()
	if err != nil {
		return nil, err
	}

	cc := make([]string, 0, len(cfg.Contexts))
	for n := range cfg.Contexts {
		cc = append(cc, n)
	}
	return cc, nil
}

// ClusterNameFromContext returns the cluster associated with the given context.
func (c *Config) ClusterNameFromContext(context string) (string, error) {
	cfg, err := c.rawConfig()
	if err != nil {
		return "", err
	}

	if ctx, ok := cfg.Contexts[context]; ok {
		return ctx.Cluster, nil
	}
	return "", fmt.Errorf("unable to locate cluster from context %s", context)
}

// CurrentClusterName returns the active cluster name.
func (c *Config) CurrentClusterName() (string, error) {
	if isSet(c.flags.ClusterName) {
		return *c.flags.ClusterName, nil
	}
	cfg, err := c.rawConfig()
	if err != nil {
		return "", err
	}
	context, err := c.CurrentContextName()
	if err != nil {
		context = cfg.CurrentContext
	}

	if ctx, ok := cfg.Contexts[context]; ok {
		return ctx.Cluster, nil
	}

	return "", errors.New("unable to locate current cluster")
}

// ClusterNames fetch all kubeconfig defined clusters.
func (c *Config) ClusterNames() ([]string, error) {
	cfg, err := c.rawConfig()
	if err != nil {
		return nil, err
	}

	cc := make([]string, 0, len(cfg.Clusters))
	for name := range cfg.Clusters {
		cc = append(cc, name)
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

	cfg, err := c.rawConfig()
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
	ns, _, err := c.clientConfig().Namespace()

	return ns, err
}

// NamespaceNames fetch all available namespaces on current cluster.
func (c *Config) NamespaceNames(nns []v1.Namespace) []string {
	nn := make([]string, 0, len(nns))
	for _, ns := range nns {
		nn = append(nn, ns.Name)
	}

	return nn
}

// ConfigAccess return the current kubeconfig api server access configuration.
func (c *Config) ConfigAccess() (clientcmd.ConfigAccess, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return c.clientConfig().ConfigAccess(), nil
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && len(*s) != 0
}

func areSet(s *[]string) bool {
	return s != nil && len(*s) != 0
}
