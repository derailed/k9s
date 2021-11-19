package config

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/derailed/k9s/internal/client"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/core/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// K9sConfig represents K9s configuration dir env var.
const K9sConfig = "K9SCONFIG"

var (
	// K9sConfigFile represents K9s config file location.
	K9sConfigFile = filepath.Join(K9sHome(), "config.yml")
	// K9sLogs represents K9s log.
	K9sLogs = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-%s.log", MustK9sUser()))
	// k9sDumpDirPath represents a base directory where K9s screen dumps will be persisted.
	k9sDumpDirPath = GetDumpConfigDirPath(K9sConfigFile)
	// K9sDumpDir represents a directory where K9s screen dumps will be persisted.
	K9sDumpDir = filepath.Join(k9sDumpDirPath, fmt.Sprintf("k9s-screens-%s", MustK9sUser()))
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
		K9s        *K9s `yaml:"k9s"`
		client     client.Connection
		settings   KubeSettings
		overrideNS bool
	}
)

// K9sHome returns k9s configs home directory.
func K9sHome() string {
	if env := os.Getenv(K9sConfig); env != "" {
		return env
	}
	if env := os.Getenv("XDG_CONFIG_HOME"); env == "" {
		dir, err := os.UserHomeDir()
		if err != nil {
			log.Error().Err(err).Msgf("user home dir")
			return ""
		}
		return path.Join(dir, ".config", "k9s")
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
	c.K9s.ActivateCluster()

	var ns string
	var override bool
	if k9sFlags != nil && IsBoolSet(k9sFlags.AllNamespaces) {
		ns, override = client.NamespaceAll, true
	} else if isSet(flags.Namespace) {
		ns, override = *flags.Namespace, true
	} else if context.Namespace != "" {
		ns = context.Namespace
	}

	if ns != "" {
		if err := c.SetActiveNamespace(ns); err != nil {
			return err
		}
		flags.Namespace, c.overrideNS = &ns, override
	}

	if isSet(flags.ClusterName) {
		c.K9s.CurrentCluster = *flags.ClusterName
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
	if c.K9s.Clusters == nil {
		log.Warn().Msgf("No context detected returning default namespace")
		return "default"
	}
	cl := c.CurrentCluster()
	if cl == nil {
		cl = NewCluster()
		c.K9s.Clusters[c.K9s.CurrentCluster] = cl
	}
	if cl.Namespace != nil {
		return cl.Namespace.Active
	}

	return "default"
}

// ValidateFavorites ensure favorite ns are legit.
func (c *Config) ValidateFavorites() {
	cl := c.K9s.ActiveCluster()
	if cl == nil {
		cl = NewCluster()
	}
	cl.Validate(c.client, c.settings)
	cl.Namespace.Validate(c.client, c.settings)
}

// FavNamespaces returns fav namespaces in the current cluster.
func (c *Config) FavNamespaces() []string {
	cl := c.K9s.ActiveCluster()
	if cl == nil {
		return nil
	}
	return c.K9s.ActiveCluster().Namespace.Favorites
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

// SetActiveView set the currently cluster active view.
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
	if c.client != nil && c.client.Config() != nil {
		c.client.Config().OverrideNS = c.overrideNS
	}
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
	EnsurePath(path, DefaultDirMod)
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
		log.Debug().Msgf("K9s cluster: %s -- %s\n", k, cl.Namespace)
	}
}

// InstantiateK9sDumpDir create path to the K9sDumpDir with configurable path
func (c *Config) InstantiateK9sDumpDir() {
	dirPath := c.K9s.GetDumpDirPath()

	if dirPath != "" && dirPath != k9sDumpDirPath {
		K9sDumpDir = filepath.Join(dirPath, fmt.Sprintf("k9s-screens-%s", MustK9sUser()))
	}

	EnsurePath(K9sDumpDir, DefaultDirMod)
}

// GetDumpConfigDirPath get dump config dir path default or from K9sConfigFile configuration. For display
func GetDumpConfigDirPath(configFilePath string) string {
	defaultDir := os.TempDir()
	if configFilePath != "" {
		f, err := os.ReadFile(configFilePath)
		if err != nil {
			return defaultDir
		}

		var cfg Config
		if err := yaml.Unmarshal(f, &cfg); err != nil {
			return defaultDir
		}
		if cfg.K9s.DumpDirPath == "" {
			return defaultDir
		}
		return cfg.K9s.DumpDirPath
	}

	return defaultDir
}

// ----------------------------------------------------------------------------
// Helpers...

func isSet(s *string) bool {
	return s != nil && len(*s) > 0
}
