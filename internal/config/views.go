// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/config/data"
	"github.com/derailed/k9s/internal/config/json"

	"gopkg.in/yaml.v2"
)

// ViewConfigListener represents a view config listener.
type ViewConfigListener interface {
	// ViewSettingsChanged notifies listener the view configuration changed.
	ViewSettingsChanged(ViewSetting)
}

// ViewSetting represents a view configuration.
type ViewSetting struct {
	Columns    []string `yaml:"columns"`
	SortColumn string   `yaml:"sortColumn"`
}

// CustomView represents a collection of view customization.
type CustomView struct {
	Views     map[string]ViewSetting `yaml:"views"`
	listeners map[string]ViewConfigListener
}

// NewCustomView returns a views configuration.
func NewCustomView() *CustomView {
	return &CustomView{
		Views:     make(map[string]ViewSetting),
		listeners: make(map[string]ViewConfigListener),
	}
}

// Reset clears out configurations.
func (v *CustomView) Reset() {
	for k := range v.Views {
		delete(v.Views, k)
	}
}

// Load loads view configurations.
func (v *CustomView) Load(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return nil
	}
	bb, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	if err := data.JSONValidator.Validate(json.ViewsSchema, bb); err != nil {
		return fmt.Errorf("validation failed for %q: %w", path, err)
	}
	var in CustomView
	if err := yaml.Unmarshal(bb, &in); err != nil {
		return err
	}
	v.Views = in.Views
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
		if v, ok := v.Views[gvr]; ok {
			list.ViewSettingsChanged(v)
		} else {
			list.ViewSettingsChanged(ViewSetting{})
		}
	}
}
