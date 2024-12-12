package config

import (
	"github.com/derailed/k9s/internal/client"
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
		"apps/v1/replicasets": {
			Name:      "apps/v1/replicasets",
			Readiness: &GVRReadiness{CellName: "Current", CellExtraName: "Desired"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellDesiredName: "Desired", CellCurrentName: "Current"},
			},
		},
		"v1/serviceaccounts":                   {Name: "v1/serviceaccounts"},
		"v1/persistentvolumeclaims":            {Name: "v1/persistentvolumeclaims"},
		"scheduling.k8s.io/v1/priorityclasses": {Name: "scheduling.k8s.io/v1/priorityclasses"},
		"v1/configmaps":                        {Name: "v1/configmaps"},
		"v1/secrets":                           {Name: "v1/secrets"},
		"v1/services":                          {Name: "v1/services"},
		"apps/v1/daemonsets": {
			Name:      "apps/v1/daemonsets",
			Readiness: &GVRReadiness{CellName: "Ready", CellExtraName: "Desired"},
			Validity: &GVRValidity{
				Replicas: Replicas{CellDesiredName: "Desired", CellCurrentName: "Ready"},
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
		"apps/v1/deployments": {
			Name:      "apps/v1/deployments",
			Readiness: &GVRReadiness{CellName: "Ready"},
			Validity: &GVRValidity{
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

// NewWorkloadGVRs returns the default GVRs to use if no custom config is set
func NewWorkloadGVRs() []WorkloadGVR {
	defaultWorkloadGVRs := make([]WorkloadGVR, 0)
	for _, gvr := range defaultConfigGVRs {
		defaultWorkloadGVRs = append(defaultWorkloadGVRs, gvr)
	}

	return defaultWorkloadGVRs
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
