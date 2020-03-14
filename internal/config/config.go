package config

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

var (
	// K9sHome represent K9s home directory.
	K9sHome = filepath.Join(mustK9sHome(), ".k9s")
	// K9sConfigFile represents K9s config file location.
	K9sConfigFile = filepath.Join(K9sHome, "config.yml")
	// K9sLogs represents K9s log.
	K9sLogs = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-%s.log", MustK9sUser()))
	// K9sDumpDir represents a directory where K9s screen dumps will be persisted.
	K9sDumpDir = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-screens-%s", MustK9sUser()))
)

type (
	// KubeSettings exposes kubeconfig context information.
	KubeSettings interface {
		// CurrentContextName returns the name of the current context.
		CurrentContextName() (string, error)

		// CurrentClusterName returns the name of the current cluster.
		CurrentClusterName() (string, error)

		// CurrentNamespace returns the name of the current namespace.
		CurrentNamespaceName() (string, error)

		// ClusterNames() returns all available cluster names.
		ClusterNames() ([]string, error)

		// NamespaceNames returns all available namespace names.
		NamespaceNames(nn []v1.Namespace) []string
	}

	// Config tracks K9s configuration options.
	Config struct {
		K9s      *K9s `yaml:"k9s"`
		client   client.Connection
		settings KubeSettings
		demoMode bool
	}
)

// NewConfig creates a new default config.
func NewConfig(ks KubeSettings) *Config {
	return &Config{K9s: NewK9s(), settings: ks}
}

// DemoMode returns true if demo mode is active, false otherwise.
func (c *Config) DemoMode() bool {
	return c.demoMode
}

// SetDemoMode sets the demo mode.
func (c *Config) SetDemoMode(b bool) {
	c.demoMode = b
}

// Refine the configuration based on cli args.
func (c *Config) Refine(flags *genericclioptions.ConfigFlags) error {
	cfg, err := flags.ToRawKubeConfigLoader().RawConfig()
	if err != nil {
		return err
	}

	if isSet(flags.Context) {
		c.K9s.CurrentContext = *flags.Context
	} else {
		c.K9s.CurrentContext = cfg.CurrentContext
	}
	log.Debug().Msgf("Active Context %q", c.K9s.CurrentContext)
	if c.K9s.CurrentContext == "" {
		return errors.New("Invalid kubeconfig context detected")
	}
	ctx, ok := cfg.Contexts[c.K9s.CurrentContext]
	if !ok {
		return fmt.Errorf("The specified context %q does not exists in kubeconfig", c.K9s.CurrentContext)
	}
	c.K9s.CurrentCluster = ctx.Cluster
	if len(ctx.Namespace) != 0 {
		if err := c.SetActiveNamespace(ctx.Namespace); err != nil {
			return err
		}
	}

	if isSet(flags.ClusterName) {
		c.K9s.CurrentCluster = *flags.ClusterName
	}

	if isSet(flags.Namespace) {
		if err := c.SetActiveNamespace(*flags.Namespace); err != nil {
			return err
		}
	}

	return nil
}

// Reset the context to the new current context/cluster.
// if it does not exist.
func (c *Config) Reset() {
	c.K9s.CurrentContext, c.K9s.CurrentCluster = "", ""
}

// CurrentCluster fetch the configuration activeCluster.
func (c *Config) CurrentCluster() *Cluster {
	if c, ok := c.K9s.Clusters[c.K9s.CurrentCluster]; ok {
		return c
	}
	return nil
}

// ActiveNamespace returns the active namespace in the current cluster.
func (c *Config) ActiveNamespace() string {
	if cl := c.CurrentCluster(); cl != nil {
		if cl.Namespace != nil {
			return cl.Namespace.Active
		}
	}
	return "default"
}

// FavNamespaces returns fav namespaces in the current cluster.
func (c *Config) FavNamespaces() []string {
	cl := c.K9s.ActiveCluster()
	if cl != nil {
		return c.K9s.ActiveCluster().Namespace.Favorites
	}
	return []string{}
}

// SetActiveNamespace set the active namespace in the current cluster.
func (c *Config) SetActiveNamespace(ns string) error {
	if c.K9s.ActiveCluster() != nil {
		return c.K9s.ActiveCluster().Namespace.SetActive(ns, c.settings)
	}

	err := errors.New("no active cluster. unable to set active namespace")
	log.Error().Err(err).Msg("SetActiveNamespace")
	return err
}

// ActiveView returns the active view in the current cluster.
func (c *Config) ActiveView() string {
	if c.K9s.ActiveCluster() == nil {
		return defaultView
	}

	cmd := c.K9s.ActiveCluster().View.Active
	if c.K9s.manualCommand != nil && *c.K9s.manualCommand != "" {
		cmd = *c.K9s.manualCommand
	}

	return cmd
}

// SetActiveView set the currently cluster active view
func (c *Config) SetActiveView(view string) {
	cl := c.K9s.ActiveCluster()
	if cl != nil {
		cl.View.Active = view
	}
}

// GetConnection return an api server connection.
func (c *Config) GetConnection() client.Connection {
	return c.client
}

// SetConnection set an api server connection.
func (c *Config) SetConnection(conn client.Connection) {
	c.client = conn
}

// Load K9s configuration from file
func (c *Config) Load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	c.K9s = NewK9s()

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		return err
	}
	if cfg.K9s != nil {
		c.K9s = cfg.K9s
	}
	if c.K9s.Logger == nil {
		c.K9s.Logger = NewLogger()
	}
	return nil
}

// Save configuration to disk.
func (c *Config) Save() error {
	log.Debug().Msg("[Config] Saving configuration...")
	c.Validate()

	return c.SaveFile(K9sConfigFile)
}

// SaveFile K9s configuration to disk.
func (c *Config) SaveFile(path string) error {
	EnsurePath(path, DefaultDirMod)
	cfg, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Msgf("[Config] Unable to save K9s config file: %v", err)
		return err
	}
	return ioutil.WriteFile(path, cfg, 0644)
}

// Validate the configuration.
func (c *Config) Validate() {
	c.K9s.Validate(c.client, c.settings)
}

// Dump debug...
func (c *Config) Dump(msg string) {
	log.Debug().Msgf("Current Cluster: %s\n", c.K9s.CurrentCluster)
	for k, cl := range c.K9s.Clusters {
		log.Debug().Msgf("K9s cluster: %s -- %s\n", k, cl.Namespace)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && len(*s) > 0
}
