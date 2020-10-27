package config

import (
	"io/ioutil"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// K9sModelConfigFile represents the location for the models configuration.
var K9sModelConfigFile = filepath.Join(K9sHome(), "models.yml")

// ModelConfigListener represents a view config listener.
type ModelConfigListener interface {
	// ModelSettingsChanged notifies listener the model configuration changed.
	ModelSettingsChanged(ModelSetting)
}

// ModelSetting represents a model configuration.
type ModelSetting struct {
	Columns []ModelColumn `yaml:"columns"`
}

// ModelSetting represents a model configuration for a single column.
type ModelColumn struct {
	Name      string `yaml:"name"`
	FieldPath string `yaml:"path"`
}

// ModelSettings represent a collection of model configurations.
type ModelSettings struct {
	Models map[string]ModelSetting `yaml:"models"`
}

// NewModelSettings returns a new configuration.
func NewModelSettings() ModelSettings {
	return ModelSettings{
		Models: make(map[string]ModelSetting),
	}
}

// CustomModel represents a collection of view customization.
type CustomModel struct {
	K9s       ModelSettings `yaml:"k9s"`
	listeners map[string]ModelConfigListener
}

// NewCustomView returns a views configuration.
func NewCustomModel() *CustomModel {
	return &CustomModel{
		K9s:       NewModelSettings(),
		listeners: make(map[string]ModelConfigListener),
	}
}

// Reset clears out configurations.
func (v *CustomModel) Reset() {
	for k := range v.K9s.Models {
		delete(v.K9s.Models, k)
	}
}

// Load loads view configurations.
func (v *CustomModel) Load(path string) error {
	raw, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}

	var in CustomModel
	if err := yaml.Unmarshal(raw, &in); err != nil {
		return err
	}
	v.K9s = in.K9s
	v.fireConfigChanged()

	return nil
}

// AddListener registers a new listener.
func (v *CustomModel) AddListener(gvr string, l ModelConfigListener) {
	v.listeners[gvr] = l
	v.fireConfigChanged()
}

// RemoveListener unregister a listener.
func (v *CustomModel) RemoveListener(gvr string) {
	delete(v.listeners, gvr)

}

func (v *CustomModel) fireConfigChanged() {
	for gvr, list := range v.listeners {
		if v, ok := v.K9s.Models[gvr]; ok {
			list.ModelSettingsChanged(v)
		}
	}
}
