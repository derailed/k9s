package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var K9sViewConfigFile = filepath.Join(K9sHome, "views_config.yml")

// ViewConfigListener represents a view config listener.
type ViewConfigListener interface {
	// ConfigChanged notifies listener the view configuration changed.
	ViewSettingsChanged(ViewSetting)
}

type ViewSetting struct {
	Columns []string `yaml:"columns"`
}

type ViewSettings struct {
	Fuck  string                 `yaml:"fuck"`
	Views map[string]ViewSetting `yaml:"views"`
}

func NewViewSettings() ViewSettings {
	return ViewSettings{
		Views: make(map[string]ViewSetting),
	}
}

// CustomView represents a collection of view customization.
type CustomView struct {
	K9s       ViewSettings `yaml:"k9s"`
	listeners map[string]ViewConfigListener
}

func NewCustomView() *CustomView {
	return &CustomView{
		K9s:       NewViewSettings(),
		listeners: make(map[string]ViewConfigListener),
	}
}

// Reset clears out configurations.
func (v *CustomView) Reset() {
	for k := range v.K9s.Views {
		delete(v.K9s.Views, k)
	}
}

// Load loads view configurations.
func (v *CustomView) Load(path string) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var in CustomView
	if err := yaml.Unmarshal(raw, &in); err != nil {
		return err
	}
	v.K9s = in.K9s
	v.fireConfigChanged()

	return nil
}

// AddListener registers a new listener.
func (c *CustomView) AddListener(gvr string, l ViewConfigListener) {
	c.listeners[gvr] = l
	c.fireConfigChanged()
}

// RemoveListener unregister a listener.
func (c *CustomView) RemoveListener(gvr string) {
	delete(c.listeners, gvr)

}

func (c *CustomView) fireConfigChanged() {
	for gvr, list := range c.listeners {
		if v, ok := c.K9s.Views[gvr]; ok {
			list.ViewSettingsChanged(v)
		}
	}
}
