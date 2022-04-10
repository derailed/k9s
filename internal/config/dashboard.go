package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var DashboarFilePath = filepath.Join(K9sHome(), "dashboard.yml")

type Dashboard struct {
	GVRs map[string]DashboardGVR `yaml:"gvrs"`
}

type DashboardGVR struct {
	Active  bool                 `yaml:"active"`
	Colors  DashboardGVRColorMap `yaml:"colors"`
	Columns DashboardGVRColumns  `yaml:"columns"`
}

type DashboardGVRColumns map[string]string

type DashboardGVRColorMap struct {
	Modified  string `yaml:"modified"`
	Added     string `yaml:"added"`
	Pending   string `yaml:"pending"`
	Error     string `yaml:"error"`
	Std       string `yaml:"std"`
	Highlight string `yaml:"highlight"`
	Kill      string `yaml:"kill"`
	Completed string `yaml:"completed"`
}

func LoadDashboard() (Dashboard, error) {
	dashboard := Dashboard{}
	f, err := os.ReadFile(DashboarFilePath)
	if err != nil {
		return dashboard, err
	}

	var config Dashboard
	if err := yaml.Unmarshal(f, &config); err != nil {
		return dashboard, err
	}

	dashboard.GVRs = config.GVRs

	return dashboard, nil
}
