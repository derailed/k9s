// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

import (
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/derailed/k9s/internal/client"
	"gopkg.in/yaml.v2"
)

var (
	// Template represents the template of a new workload gvr
	Template = []byte(`name: "test.com/v1alpha1/myCRD"
status:
  cellName: "Status"
  # na: true
readiness:
  cellName: "Current"
  # The cellExtraName will be shown as cellName/cellExtraName
  cellExtraName: "Desired"
  # na: true
validity:
  replicas:
    cellCurrentName: "Current"
    cellDesiredName: "Desired"
    # cellAllName: "Ready"
  matchs:
    - cellName: "State"
      cellValue: "Ready"`)
)

var (
	// defaultGvr represent the default values uses if a custom gvr is set without status, validity or readiness
	defaultGvr = WorkloadGVR{
		Status:    &GVRStatus{CellName: "Status"},
		Validity:  &GVRValidity{Matchs: []Match{{CellName: "Ready", Value: "True"}}},
		Readiness: &GVRReadiness{CellName: "Ready"},
	}

	// defaultConfigGVRs represents the default configurations
	defaultConfigGVRs = map[string]WorkloadGVR{
		"apps/v1/deployments": {
			Name:      "apps/v1/deployments",
			Readiness: &GVRReadiness{CellName: "Ready"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellAllName: "Ready"},
			},
		},
		"apps/v1/daemonsets": {
			Name:      "apps/v1/daemonsets",
			Readiness: &GVRReadiness{CellName: "Ready", CellExtraName: "Desired"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellDesiredName: "Desired", CellCurrentName: "Ready"},
			},
		},
		"apps/v1/replicasets": {
			Name:      "apps/v1/replicasets",
			Readiness: &GVRReadiness{CellName: "Current", CellExtraName: "Desired"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellDesiredName: "Desired", CellCurrentName: "Current"},
			},
		},
		"apps/v1/statefulSets": {
			Name:      "apps/v1/statefulSets",
			Status:    &GVRStatus{CellName: "Ready"},
			Readiness: &GVRReadiness{CellName: "Ready"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellAllName: "Ready"},
			},
		},
		"v1/pods": {
			Name:      "v1/pods",
			Status:    &GVRStatus{CellName: "Status"},
			Readiness: &GVRReadiness{CellName: "Ready"},
			Validity: &GVRValidity{
				Matchs: []Match{
					{CellName: "Status", Value: "Running"},
				},
				Replicas: Replicas{CellAllName: "Ready"},
			},
		},
	}
)

type CellName string

type GVRStatus struct {
	NA       bool     `json:"na" yaml:"na"`
	CellName CellName `json:"cellName" yaml:"cellName"`
}

type GVRReadiness struct {
	NA            bool     `json:"na" yaml:"na"`
	CellName      CellName `json:"cellName" yaml:"cellName"`
	CellExtraName CellName `json:"cellExtraName" yaml:"cellExtraName"`
}

type Match struct {
	CellName CellName `json:"cellName" yaml:"cellName"`
	Value    string   `json:"cellValue" yaml:"cellValue"`
}

type Replicas struct {
	CellCurrentName CellName `json:"cellCurrentName" yaml:"cellCurrentName"`
	CellDesiredName CellName `json:"cellDesiredName" yaml:"cellDesiredName"`
	CellAllName     CellName `json:"cellAllName" yaml:"cellAllName"`
}

type GVRValidity struct {
	NA       bool     `json:"na" yaml:"na"`
	Matchs   []Match  `json:"matchs,omitempty" yaml:"matchs,omitempty"`
	Replicas Replicas `json:"replicas" yaml:"replicas"`
}

type WorkloadGVR struct {
	Name      string        `json:"name" yaml:"name"`
	Status    *GVRStatus    `json:"status,omitempty" yaml:"status,omitempty"`
	Readiness *GVRReadiness `json:"readiness,omitempty" yaml:"readiness,omitempty"`
	Validity  *GVRValidity  `json:"validity,omitempty" yaml:"validity,omitempty"`
}

type WorkloadConfig struct {
	GVRFilenames []string `yaml:"wkg"`
}

// NewWorkloadGVRs returns the default GVRs to use if no custom config is set
// The workloadDir represent the directory of the custom workloads, the gvrNames are the custom gvrs names
func NewWorkloadGVRs(workloadDir string, gvrNames []string) ([]WorkloadGVR, error) {
	workloadGVRs := make([]WorkloadGVR, 0)
	for _, gvr := range defaultConfigGVRs {
		workloadGVRs = append(workloadGVRs, gvr)
	}

	var errs error

	// Append custom GVRS
	if len(gvrNames) != 0 {
		for _, filename := range gvrNames {
			wkgvr, err := GetWorkloadGVRFromFile(path.Join(workloadDir, fmt.Sprintf("%s.%s", filename, "yaml")))
			if err != nil {
				errs = errors.Join(errs, err)
				continue
			}
			workloadGVRs = append(workloadGVRs, wkgvr)
		}
	}

	return workloadGVRs, errs
}

// GetWorkloadGVRFromFile returns a gvr from a filepath
func GetWorkloadGVRFromFile(filepath string) (WorkloadGVR, error) {
	yamlFile, err := os.ReadFile(filepath)
	if err != nil {
		return WorkloadGVR{}, err
	}

	var wkgvr WorkloadGVR
	if err = yaml.Unmarshal(yamlFile, &wkgvr); err != nil {
		return WorkloadGVR{}, err
	}

	return wkgvr, nil
}

// GetGVR will return the GVR defined by the WorkloadGVR's name
func (wgvr WorkloadGVR) GetGVR() client.GVR {
	return client.NewGVR(wgvr.Name)
}

// ApplyDefault will complete the GVR with missing values
// If it's an existing GVR's name, it will apply their corresponding default values
// If it's an unknown resources without readiness, status or validity it will use the default ones
func (wkgvr *WorkloadGVR) ApplyDefault() {
	// Apply default values
	existingGvr, ok := defaultConfigGVRs[wkgvr.Name]
	if ok {
		wkgvr.applyDefaultValues(existingGvr)
	} else {
		wkgvr.applyDefaultValues(defaultGvr)
	}
}

func (wkgvr *WorkloadGVR) applyDefaultValues(defaultGVR WorkloadGVR) {
	if wkgvr.Status == nil {
		wkgvr.Status = defaultGVR.Status
	}

	if wkgvr.Readiness == nil {
		wkgvr.Readiness = defaultGVR.Readiness
	}

	if wkgvr.Validity == nil {
		wkgvr.Validity = defaultGVR.Validity
	}
}
