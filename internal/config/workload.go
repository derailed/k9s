package config

import "github.com/derailed/k9s/internal/client"

var (
	// TODO: Remove that and add it to the doc (with basic example)
	// yaml
	// customWorkloadGVRs:
	// 	- Name: ""
	// 	Status:
	//		CellName: ""
	//	Readiness:
	// 		CellName: ""
	// 		ExtraCellName: ""
	//	Validity:
	//		Replicas:
	// 			AllCellName: ""
	// 			CurrentCellName: ""
	// 			DesiredCellName: ""
	// 		Matchs:
	//			- CellName: ""
	//			  Value : ""
	//			- CellName: ""
	//			  Value: ""
	//				...

	defaultGVRs = map[string]WorkloadGVR{
		"v1/pods": {
			Name:      "v1/pods",
			Status:    &WKStatus{CellName: "Status"},
			Readiness: &Readiness{CellName: "Ready"},
			Validity: &Validity{
				Matchs: []Match{
					{CellName: "Status", Value: "Running"},
				},
				Replicas: Replicas{AllCellName: "Ready"},
			},
		},
		"apps/v1/replicasets": {
			Name:      "apps/v1/replicasets",
			Readiness: &Readiness{CellName: "Current", ExtraCellName: "Desired"},
			Validity: &Validity{
				Replicas: Replicas{DesiredCellName: "Desired", CurrentCellName: "Current"},
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
			Readiness: &Readiness{CellName: "Ready", ExtraCellName: "Desired"},
			Validity: &Validity{
				Replicas: Replicas{DesiredCellName: "Desired", CurrentCellName: "Ready"},
			},
		},
		"apps/v1/statefulSets": {
			Name:      "apps/v1/statefulSets",
			Status:    &WKStatus{CellName: "Ready"},
			Readiness: &Readiness{CellName: "Ready"},
			Validity: &Validity{
				Replicas: Replicas{AllCellName: "Ready"},
			},
		},
		"apps/v1/deployments": {
			Name:      "apps/v1/deployments",
			Readiness: &Readiness{CellName: "Ready"},
			Validity: &Validity{
				Replicas: Replicas{AllCellName: "Ready"},
			},
		},
	}
)

// TODO: Rename all fields with better names

type CellName string

type WKStatus struct {
	CellName CellName `json:"name" yaml:"name"`
}

type Readiness struct {
	CellName      CellName `json:"name" yaml:"name"`
	ExtraCellName CellName `json:"extra_cell_name" yaml:"extra_cell_name"`
}

type Match struct {
	CellName CellName `json:"name" yaml:"name"`
	Value    string   `json:"value" yaml:"value"`
}

type Replicas struct {
	CurrentCellName CellName `json:"currentName" yaml:"currentName"`
	DesiredCellName CellName `json:"desiredName" yaml:"desiredName"`
	AllCellName     CellName `json:"allName" yaml:"allName"`
}

type Validity struct {
	Matchs   []Match  `json:"matchs,omitempty" yaml:"matchs,omitempty"`
	Replicas Replicas `json:"replicas" yaml:"replicas"`
}

type WorkloadGVR struct {
	Name      string     `json:"name" yaml:"name"`
	Status    *WKStatus  `json:"status,omitempty" yaml:"status,omitempty"`
	Readiness *Readiness `json:"readiness,omitempty" yaml:"readiness,omitempty"`
	Validity  *Validity  `json:"validity,omitempty" yaml:"validity,omitempty"`
}

// TODO: Find a better name, this only create the default gvr values
func NewDefaultWorkloadGVRs() []WorkloadGVR {
	defaultWorkloadGVRs := make([]WorkloadGVR, 0)
	for _, gvr := range defaultGVRs {
		defaultWorkloadGVRs = append(defaultWorkloadGVRs, gvr)
	}

	return defaultWorkloadGVRs
}

// TODO: Add comment
func (wgvr WorkloadGVR) GetGVR() client.GVR {
	return client.NewGVR(wgvr.Name)
}

// TODO: Add comment, this is applying default value for GVR set partially
func (wkgvr *WorkloadGVR) ApplyDefault() {
	if existingGvr, ok := defaultGVRs[wkgvr.Name]; ok {
		if wkgvr.Status == nil {
			wkgvr.Status = existingGvr.Status
		}
		if wkgvr.Readiness == nil {
			wkgvr.Readiness = existingGvr.Readiness
		}
		if wkgvr.Validity == nil {
			wkgvr.Validity = existingGvr.Validity
		}
	}
}
