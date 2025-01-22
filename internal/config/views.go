// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"cmp"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"slices"
	"strings"

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

func (v *ViewSetting) HasCols() bool {
	return len(v.Columns) > 0
}

func (v *ViewSetting) IsBlank() bool {
	return v == nil || len(v.Columns) == 0
}

func (v *ViewSetting) SortCol() (string, bool, error) {
	if v == nil || v.SortColumn == "" {
		return "", false, fmt.Errorf("no sort column specified")
	}
	tt := strings.Split(v.SortColumn, ":")
	if len(tt) < 2 {
		return "", false, fmt.Errorf("invalid sort column spec: %q. must be col-name:asc|desc", v.SortColumn)
	}

	return tt[0], tt[1] == "asc", nil
}

func (v *ViewSetting) Equals(vs *ViewSetting) bool {
	if v == nil || vs == nil {
		return v == nil && vs == nil
	}
	if c := slices.Compare(v.Columns, vs.Columns); c != 0 {
		return false
	}
	return cmp.Compare(v.SortColumn, vs.SortColumn) == 0
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
	if _, err := os.Stat(path); errors.Is(err, fs.ErrNotExist) {
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
		if view, ok := v.Views[gvr]; ok {
			list.ViewSettingsChanged(view)
		} else {
			list.ViewSettingsChanged(ViewSetting{})
		}
	}
}
