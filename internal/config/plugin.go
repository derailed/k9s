package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
)

// K9sPluginsFilePath manages K9s plugins.
var K9sPluginsFilePath = filepath.Join(K9sHome(), "plugin.yml")
var K9sPluginDirectory = filepath.Join("k9s", "plugins")

// Plugins represents a collection of plugins.
type Plugins struct {
	Plugin map[string]Plugin `yaml:"plugin"`
}

// Plugin describes a K9s plugin.
type Plugin struct {
	Scopes      []string `yaml:"scopes"`
	Args        []string `yaml:"args"`
	ShortCut    string   `yaml:"shortCut"`
	Pipes       []string `yaml:"pipes"`
	Description string   `yaml:"description"`
	Command     string   `yaml:"command"`
	Confirm     bool     `yaml:"confirm"`
	Background  bool     `yaml:"background"`
}

func (p Plugin) String() string {
	return fmt.Sprintf("[%s] %s(%s)", p.ShortCut, p.Command, strings.Join(p.Args, " "))
}

// NewPlugins returns a new plugin.
func NewPlugins() Plugins {
	return Plugins{
		Plugin: make(map[string]Plugin),
	}
}

// Load K9s plugins.
func (p Plugins) Load() error {
	var pluginDirs []string
	for _, dataDir := range xdg.DataDirs {
		pluginDirs = append(pluginDirs, filepath.Join(dataDir, K9sPluginDirectory))
	}
	return p.LoadPlugins(K9sPluginsFilePath, pluginDirs)
}

// LoadPlugins loads plugins from a given file and a set of plugin directories.
func (p Plugins) LoadPlugins(path string, pluginDirs []string) error {
	f, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var pp Plugins
	if err := yaml.Unmarshal(f, &pp); err != nil {
		return err
	}

	for _, pluginDir := range pluginDirs {
		pluginFiles, err := os.ReadDir(pluginDir)
		if err != nil {
			log.Warn().Msgf("Failed reading plugin path %s; %s", pluginDir, err)
			continue
		}
		for _, file := range pluginFiles {
			if file.IsDir() || !isYamlFile(file) {
				continue
			}
			pluginFile, err := os.ReadFile(filepath.Join(pluginDir, file.Name()))
			if err != nil {
				return err
			}
			var plugin Plugin
			if err = yaml.Unmarshal(pluginFile, &plugin); err != nil {
				return err
			}
			p.Plugin[strings.TrimSuffix(file.Name(), filepath.Ext(file.Name()))] = plugin
		}
	}

	for k, v := range pp.Plugin {
		p.Plugin[k] = v
	}

	return nil
}

func isYamlFile(file os.DirEntry) bool {
	ext := filepath.Ext(file.Name())
	return ext == ".yml" || ext == ".yaml"
}
