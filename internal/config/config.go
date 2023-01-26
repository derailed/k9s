package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// K9sConfig represents K9s configuration dir env var.
const K9sConfig = "K9SCONFIG"

var (
	// K9sConfigFile represents K9s config file location.
	K9sConfigFile = filepath.Join(K9sHome(), "config.yml")
	// K9sDefaultScreenDumpDir represents a default directory where K9s screen dumps will be persisted.
	K9sDefaultScreenDumpDir = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-screens-%s", MustK9sUser()))
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
		ClusterNames() (map[string]struct{}, error)
	}

	// Config tracks K9s configuration options.
	Config struct {
		K9s      *K9s `yaml:"k9s"`
		client   client.Connection
		settings KubeSettings
	}
)

// K9sHome returns k9s configs home directory.
func K9sHome() string {
	if env := os.Getenv(K9sConfig); env != "" {
		return env
	}

	xdgK9sHome, err := xdg.ConfigFile("k9s")
	if err != nil {
		log.Fatal().Err(err).Msg("Unable to create configuration directory for k9s")
	}

	return xdgK9sHome
}

// NewConfig creates a new default config.
func NewConfig(ks KubeSettings) *Config {
	return &Config{K9s: NewK9s(), settings: ks}
}

// Refine the configuration based on cli args.
func (c *Config) Refine(flags *genericclioptions.ConfigFlags, k9sFlags *Flags, cfg *client.Config) error {
	if isSet(flags.Context) {
		c.K9s.CurrentContext = *flags.Context
	} else {
		context, err := cfg.CurrentContextName()
		if err != nil {
			return err
		}
		c.K9s.CurrentContext = context
	}
	log.Debug().Msgf("Active Context %q", c.K9s.CurrentContext)
	if c.K9s.CurrentContext == "" {
		return errors.New("Invalid kubeconfig context detected")
	}
	cc, err := cfg.Contexts()
	if err != nil {
		return err
	}
	context, ok := cc[c.K9s.CurrentContext]
	if !ok {
		return fmt.Errorf("The specified context %q does not exists in kubeconfig", c.K9s.CurrentContext)
	}
	c.K9s.CurrentCluster = context.Cluster
	c.K9s.ActivateCluster(context.Namespace)

	var ns = client.DefaultNamespace
	if k9sFlags != nil && IsBoolSet(k9sFlags.AllNamespaces) {
		ns = client.NamespaceAll
	} else if isSet(flags.Namespace) {
		ns = *flags.Namespace
	} else {
		ns = context.Namespace
	}

	if err := c.SetActiveNamespace(ns); err != nil {
		return err
	}
	flags.Namespace = &ns

	if isSet(flags.ClusterName) {
		c.K9s.CurrentCluster = *flags.ClusterName
	}

	return EnsureDirPath(c.K9s.GetScreenDumpDir(), DefaultDirMod)
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
	if c.K9s.Clusters == nil {
		log.Warn().Msgf("No context detected returning default namespace")
		return "default"
	}
	cl := c.CurrentCluster()
	if cl != nil && cl.Namespace != nil {
		return cl.Namespace.Active
	}
	if cl == nil {
		cl = NewCluster()
		c.K9s.Clusters[c.K9s.CurrentCluster] = cl
	}
	if ns, err := c.settings.CurrentNamespaceName(); err == nil && ns != "" {
		if cl.Namespace == nil {
			cl.Namespace = NewNamespace()
		}
		cl.Namespace.Active = ns
		return ns
	}

	return "default"
}

// ValidateFavorites ensure favorite ns are legit.
func (c *Config) ValidateFavorites() {
	cl := c.K9s.ActiveCluster()
	cl.Validate(c.client, c.settings)
	cl.Namespace.Validate(c.client, c.settings)
}

// FavNamespaces returns fav namespaces in the current cluster.
func (c *Config) FavNamespaces() []string {
	cl := c.K9s.ActiveCluster()

	return cl.Namespace.Favorites
}

// SetActiveNamespace set the active namespace in the current cluster.
func (c *Config) SetActiveNamespace(ns string) error {
	if cl := c.K9s.ActiveCluster(); cl != nil {
		return cl.Namespace.SetActive(ns, c.settings)
	}
	err := errors.New("no active cluster. unable to set active namespace")
	log.Error().Err(err).Msg("SetActiveNamespace")

	return err
}

// ActiveView returns the active view in the current cluster.
func (c *Config) ActiveView() string {
	cl := c.K9s.ActiveCluster()
	if cl == nil {
		return defaultView
	}
	cmd := cl.View.Active
	if c.K9s.manualCommand != nil && *c.K9s.manualCommand != "" {
		cmd = *c.K9s.manualCommand
	}

	return cmd
}

// SetActiveView set the currently cluster active view.
func (c *Config) SetActiveView(view string) {
	if cl := c.K9s.ActiveCluster(); cl != nil {
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

// Load K9s configuration from file.
func (c *Config) Load(path string) error {
	f, err := os.ReadFile(path)
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
	c.Validate()

	return c.SaveFile(K9sConfigFile)
}

// SaveFile K9s configuration to disk.
func (c *Config) SaveFile(path string) error {
	if err := EnsureDirPath(path, DefaultDirMod); err != nil {
		return err
	}
	cfg, err := yaml.Marshal(c)
	if err != nil {
		log.Error().Msgf("[Config] Unable to save K9s config file: %v", err)
		return err
	}
	return os.WriteFile(path, cfg, 0644)
}

// Validate the configuration.
func (c *Config) Validate() {
	c.K9s.Validate(c.client, c.settings)
}

// Dump debug...
func (c *Config) Dump(msg string) {
	log.Debug().Msgf("Current Cluster: %s\n", c.K9s.CurrentCluster)
	for k, cl := range c.K9s.Clusters {
		log.Debug().Msgf("K9s cluster: %s -- %+v\n", k, cl.Namespace)
	}
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && len(*s) > 0
}
