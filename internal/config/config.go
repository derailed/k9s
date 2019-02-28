package config

// BOZO!! Once yaml is stable implement validation
// go get gopkg.in/validator.v2

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/derailed/k9s/internal/resource"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	// Root K9s configuration.
	Root *Config
	// K9sHome represent K9s home directory.
	K9sHome = filepath.Join(mustK9sHome(), ".k9s")
	// K9sConfigFile represents K9s config file location.
	K9sConfigFile = filepath.Join(K9sHome, "config.yml")
	// K9sLogs represents K9s log.
	K9sLogs = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-%s.log", mustK9sUser()))
)

// KubeSettings exposes kubeconfig context informations.
type KubeSettings interface {
	CurrentContextName() (string, error)
	CurrentClusterName() (string, error)
	CurrentNamespaceName() (string, error)
	ClusterNames() ([]string, error)
	NamespaceNames() ([]string, error)
}

// Config tracks K9s configuration options.
type Config struct {
	K9s      *K9s `yaml:"k9s"`
	settings KubeSettings
}

// NewConfig creates a new default config.
func NewConfig(ks KubeSettings) *Config {
	return &Config{K9s: NewK9s(), settings: ks}
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
	return resource.DefaultNamespace
}

// FavNamespaces returns fav namespaces in the current cluster.
func (c *Config) FavNamespaces() []string {
	return c.K9s.ActiveCluster().Namespace.Favorites
}

// SetActiveNamespace set the active namespace in the current cluster.
func (c *Config) SetActiveNamespace(ns string) {
	if c.K9s.ActiveCluster() != nil {
		c.K9s.ActiveCluster().Namespace.SetActive(ns)
	} else {
		log.Debug("Doh! no active cluster. unable to set active namespace")
	}
}

// ActiveView returns the active view in the current cluster.
func (c *Config) ActiveView() string {
	if c.K9s.ActiveCluster() == nil {
		return defaultView
	}
	return c.K9s.ActiveCluster().View.Active
}

// SetActiveView set the currently cluster active view
func (c *Config) SetActiveView(view string) {
	c.K9s.ActiveCluster().View.Active = view
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
	return nil
}

// Save configuration to disk.
func (c *Config) Save() error {
	log.Debugf("[Config] Saving configuration...")
	c.Validate()
	return c.SaveFile(K9sConfigFile)
}

// SaveFile K9s configuration to disk.
func (c *Config) SaveFile(path string) error {
	EnsurePath(path, DefaultDirMod)
	cfg, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("[Config] Unable to save K9s config file: %v", err)
		return err
	}
	return ioutil.WriteFile(path, cfg, 0644)
}

// Validate the configuration.
func (c *Config) Validate() {
	c.K9s.Validate(c.settings)
}

// Dump debug...
func (c *Config) Dump(msg string) {
	log.Debug(msg)
	log.Debugf("Current Context: %s\n", c.K9s.CurrentCluster)
	for k, cl := range c.K9s.Clusters {
		log.Debugf("K9s cluster: %s -- %s\n", k, cl.Namespace)
	}
}
