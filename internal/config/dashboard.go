package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

var DashboarFilePath = filepath.Join(K9sHome(), "dashboard.yml")

type Dashboard struct {
	GVRs []string `yaml:"gvrs"`
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
