package k8s

import (
	"errors"
	"fmt"

	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
	restclient "k8s.io/client-go/rest"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

const defaultNamespace = "default"

// Config tracks a kubernetes configuration.
type Config struct {
	flags          *genericclioptions.ConfigFlags
	clientConfig   clientcmd.ClientConfig
	currentContext string
	rawConfig      *clientcmdapi.Config
	restConfig     *restclient.Config
}

// NewConfig returns a new k8s config or an error if the flags are invalid.
func NewConfig(f *genericclioptions.ConfigFlags) *Config {
	return &Config{flags: f}
}

// Flags returns configuration flags.
func (c *Config) Flags() *genericclioptions.ConfigFlags {
	return c.flags
}

// SwitchContext changes the kubeconfig context to a new cluster.
func (c *Config) SwitchContext(name string) error {
	currentCtx, err := c.CurrentContextName()
	if err != nil {
		return err
	}

	if currentCtx != name {
		c.reset()
		c.flags.Context, c.currentContext = &name, name
	}
	return nil
}

func (c *Config) reset() {
	c.clientConfig, c.rawConfig, c.restConfig = nil, nil, nil
}

// CurrentContextName returns the currently active config context.
func (c *Config) CurrentContextName() (string, error) {
	if isSet(c.flags.Context) {
		return *c.flags.Context, nil
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "", err
	}
	return cfg.CurrentContext, nil
}

// GetContext fetch a given context or error if it does not exists.
func (c *Config) GetContext(n string) (*clientcmdapi.Context, error) {
	cfg, err := c.RawConfig()
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

	return clientcmd.ModifyConfig(c.clientConfig.ConfigAccess(), cfg, true)
}

// ContextNames fetch all available contexts.
func (c *Config) ContextNames() ([]string, error) {
	cfg, err := c.RawConfig()
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
func (c *Config) ClusterNameFromContext(ctx string) (string, error) {
	cfg, err := c.RawConfig()
	if err != nil {
		return "", err
	}

	if ctx, ok := cfg.Contexts[ctx]; ok {
		return ctx.Cluster, nil
	}
	return "", fmt.Errorf("unable to locate cluster from context %s", ctx)
}

// CurrentClusterName returns the active cluster name.
func (c *Config) CurrentClusterName() (string, error) {
	if isSet(c.flags.ClusterName) {
		return *c.flags.ClusterName, nil
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
		return ctx.Cluster, nil
	}

	return "", errors.New("unable to locate current cluster")
}

// ClusterNames fetch all kubeconfig defined clusters.
func (c *Config) ClusterNames() ([]string, error) {
	cfg, err := c.RawConfig()
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
	if c.flags.ImpersonateGroup != nil && len(*c.flags.ImpersonateGroup) != 0 {
		return *c.flags.ImpersonateGroup, nil
	}

	return []string{}, errors.New("unable to locate current group")
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
	if isSet(c.flags.Namespace) {
		return *c.flags.Namespace, nil
	}

	cfg, err := c.RawConfig()
	if err != nil {
		return "", err
	}
	ctx, err := c.CurrentContextName()
	if err != nil {
		return "", err
	}
	if ctx, ok := cfg.Contexts[ctx]; ok {
		if isSet(&ctx.Namespace) {
			return ctx.Namespace, nil
		}
	}

	return "", fmt.Errorf("No active namespace specified")
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
	c.ensureConfig()
	return c.clientConfig.ConfigAccess(), nil
}

// RawConfig fetch the current kubeconfig with no overrides.
func (c *Config) RawConfig() (clientcmdapi.Config, error) {
	if c.rawConfig != nil {
		if c.rawConfig.CurrentContext == c.currentContext {
			return *c.rawConfig, nil
		}
		log.Debug().Msgf("Context switch detected... %s vs %s", c.rawConfig.CurrentContext, c.currentContext)
		c.currentContext = c.rawConfig.CurrentContext
		c.reset()
	}

	if c.rawConfig == nil {
		c.ensureConfig()
		cfg, err := c.clientConfig.RawConfig()
		if err != nil {
			return cfg, err
		}
		c.rawConfig = &cfg
		c.currentContext = cfg.CurrentContext
	}

	return *c.rawConfig, nil
}

// RESTConfig fetch the current REST api service connection.
func (c *Config) RESTConfig() (*restclient.Config, error) {
	if c.restConfig != nil {
		return c.restConfig, nil
	}

	var err error
	if c.restConfig, err = c.flags.ToRESTConfig(); err != nil {
		return nil, err
	}
	log.Debug().Msgf("Connecting to API Server %s", c.restConfig.Host)

	return c.restConfig, nil
}

func (c *Config) ensureConfig() {
	if c.clientConfig != nil {
		return
	}

	log.Debug().Msg("Loading raw config from flags...")
	c.clientConfig = c.flags.ToRawKubeConfigLoader()
	return
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && len(*s) != 0
}
