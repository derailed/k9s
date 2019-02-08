package config

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v2"
)

var (
	// Root K9s configuration.
	Root = NewConfig()
	// K9sHome represent K9s home directory.
	K9sHome = filepath.Join(mustK9sHome(), ".k9s")
	// K9sConfigFile represents K9s config file location.
	K9sConfigFile = filepath.Join(K9sHome, "config.yml")
	// K9sLogs represents K9s log.
	K9sLogs = filepath.Join(os.TempDir(), fmt.Sprintf("k9s-%s.log", mustK9sUser()))
)

type ClusterInfo interface {
	ActiveClusterOrDie() string
	AllClustersOrDie() []string
	AllNamespacesOrDie() []string
}

// Config tracks K9s configuration options.
type Config struct {
	K9s *K9s `yaml:"k9s"`
}

// NewConfig creates a new default config.
func NewConfig() *Config {
	return &Config{K9s: NewK9s()}
}

// ActiveClusterName fetch the configuration activeCluster.
func (c *Config) ActiveClusterName() string {
	if c.K9s.Context == nil {
		c.K9s.Context = NewContext()
	}
	return c.K9s.Context.Active
}

// ActiveCluster fetch the configuration activeCluster.
func (c *Config) ActiveCluster() *Cluster {
	if c.K9s.Context == nil {
		c.K9s.Context = NewContext()
	}
	return c.K9s.ActiveCluster()
}

// SetActiveCluster set the active cluster to the a new configuration.
func (c *Config) SetActiveCluster(s string) {
	c.K9s.Context.SetActiveCluster(s)
}

// ActiveNamespace returns the active namespace in the current cluster.
func (c *Config) ActiveNamespace() string {
	return c.K9s.ActiveCluster().Namespace.Active
}

// FavNamespaces returns fav namespaces in the current cluster.
func (c *Config) FavNamespaces() []string {
	return c.K9s.ActiveCluster().Namespace.Favorites
}

// SetActiveNamespace set the active namespace in the current cluster.
func (c *Config) SetActiveNamespace(ns string) {
	c.K9s.ActiveCluster().Namespace.SetActive(ns)
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
	if c.K9s.Context == nil {
		c.K9s.Context = NewContext()
	}
	c.K9s.ActiveCluster().View.Active = view
}

// Load K9s configuration from file
func Load(path string) error {
	f, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var cfg Config
	if err := yaml.Unmarshal(f, &cfg); err != nil {
		Root = NewConfig()
		return err
	}
	Root = &cfg
	return nil
}

// Save configuration to disk.
func (c *Config) Save(ci ClusterInfo) error {
	c.Validate(ci)
	return c.SaveFile(K9sConfigFile)
}

// SaveFile K9s configuration to disk.
func (c *Config) SaveFile(path string) error {
	log.Debugf("[Config] Saving configuration")
	ensurePath(path, 0755)
	cfg, err := yaml.Marshal(c)
	if err != nil {
		log.Errorf("[Config] Unable to save K9s config file: %v", err)
		return err
	}
	return ioutil.WriteFile(path, cfg, 0644)
}

func (c *Config) activeCluster() *Cluster {
	return c.K9s.Context.Clusters[c.K9s.Context.Active]
}

// Validate the configuration.
func (c *Config) Validate(ci ClusterInfo) {
	if c.K9s == nil {
		c.K9s = NewK9s()
	}
	c.K9s.Validate(ci)
}
