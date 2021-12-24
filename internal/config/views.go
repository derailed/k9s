package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// K9sViewConfigFile represents the location for the views configuration.
var K9sViewConfigFile = filepath.Join(K9sHome(), "views.yml")

// ViewConfigListener represents a view config listener.
type ViewConfigListener interface {
	// ConfigChanged notifies listener the view configuration changed.
	ViewSettingsChanged(ViewSetting)
}

// ViewSetting represents a view configuration.
type ViewSetting struct {
	Columns    []string `yaml:"columns"`
	SortColumn string   `yaml:"sortColumn"`
}

// ViewSettings represent a collection of view configurations.
type ViewSettings struct {
	Views map[string]ViewSetting `yaml:"views"`
}

// NewViewSettings returns a new configuration.
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

// NewCustomView returns a views configuration.
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
	raw, err := os.ReadFile(path)
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
func (v *CustomView) AddListener(gvr string, l ViewConfigListener) {
	v.listeners[gvr] = l
	v.fireConfigChanged()
}

// RemoveListener unregister a listener.
func (v *CustomView) RemoveListener(gvr string) {
	delete(v.listeners, gvr)
}

func (v *CustomView) fireConfigChanged() {
	for gvr, list := range v.listeners {
		if v, ok := v.K9s.Views[gvr]; ok {
			list.ViewSettingsChanged(v)
		}
	}
}
