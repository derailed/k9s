// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config

type (
	// Emoji represents emojis configuration
	Emoji struct {
		StartUp            string    `json:"startUp" yaml:"startUp"`
		CommandLine        string    `json:"commandLine" yaml:"commandLine"`
		FilterLine         string    `json:"filterLine" yaml:"filterLine"`
		Happy              string    `json:"happy" yaml:"happy"`
		Warn               string    `json:"warn" yaml:"warn"`
		Angry              string    `json:"angry" yaml:"angry"`
		File               string    `json:"file" yaml:"file"`
		Folder             string    `json:"folder" yaml:"folder"`
		CheckMark          string    `json:"checkMark" yaml:"checkMark"`
		LogStreamCancelled string    `json:"logStreamCancelled" yaml:"logStreamCancelled"`
		NewVersion         string    `json:"newVersion" yaml:"newVersion"`
		XRay               XRayEmoji `json:"xray" yaml:"xray"`
	}

	// XRayEmoji represents xray view emojis
	XRayEmoji struct {
		Namespaces               string `json:"namespaces" yaml:"namespaces"`
		DefaultGvr               string `json:"defaultGvr" yaml:"defaultGvr"`
		Nodes                    string `json:"nodes" yaml:"nodes"`
		Pods                     string `json:"pods" yaml:"pods"`
		Services                 string `json:"services" yaml:"services"`
		ServiceAccounts          string `json:"serviceAccounts" yaml:"serviceAccounts"`
		PersistentVolumes        string `json:"persistentVolumes" yaml:"persistentVolumes"`
		PersistentVolumeClaims   string `json:"persistentVolumeClaims" yaml:"persistentVolumeClaims"`
		Secrets                  string `json:"secrets" yaml:"secrets"`
		HorizontalPodAutoscalers string `json:"horizontalPodAutoscalers" yaml:"horizontalPodAutoscalers"`
		ConfigMaps               string `json:"configMaps" yaml:"configMaps"`
		Deployments              string `json:"deployments" yaml:"deployments"`
		StatefulSets             string `json:"statefulSets" yaml:"statefulSets"`
		DaemonSets               string `json:"daemonSets" yaml:"daemonSets"`
		ReplicaSets              string `json:"replicaSets" yaml:"replicaSets"`
		ClusterRoles             string `json:"clusterRoles" yaml:"clusterRoles"`
		Roles                    string `json:"roles" yaml:"roles"`
		NetworkPolices           string `json:"networkPolices" yaml:"networkPolices"`
		PodDisruptionBudgets     string `json:"podDisruptionBudgets" yaml:"podDisruptionBudgets"`
		PodSecurityPolicies      string `json:"podSecurityPolicies" yaml:"podSecurityPolicies"`
		Containers               string `json:"containers" yaml:"containers"`
		Report                   string `json:"report" yaml:"report"`
		Issue0                   string `json:"issue0" yaml:"issue0"`
		Issue1                   string `json:"issue1" yaml:"issue1"`
		Issue2                   string `json:"issue2" yaml:"issue2"`
		Issue3                   string `json:"issue3" yaml:"issue3"`
	}
)

// --- Getters for main emojis ---

// StartUpEmoji returns the startup emoji
func (e *Emoji) StartUpEmoji() string {
	if e.StartUp == "" {
		return DefaultStartUpEmoji
	}
	return e.StartUp
}

// CommandLineEmoji returns the command line emoji
func (e *Emoji) CommandLineEmoji() string {
	if e.CommandLine == "" {
		return DefaultCommandLineEmoji
	}
	return e.CommandLine
}

// FilterLineEmoji returns the filter line emoji
func (e *Emoji) FilterLineEmoji() string {
	if e.FilterLine == "" {
		return DefaultFilterLineEmoji
	}
	return e.FilterLine
}

// InfoStatusEmoji returns the happy/info status emoji
func (e *Emoji) InfoStatusEmoji() string {
	if e.Happy == "" {
		return DefaultHappyEmoji
	}
	return e.Happy
}

// WarningStatusEmoji returns the warning status emoji
func (e *Emoji) WarningStatusEmoji() string {
	if e.Warn == "" {
		return DefaultWarnEmoji
	}
	return e.Warn
}

// ErrorStatusEmoji returns the error status emoji
func (e *Emoji) ErrorStatusEmoji() string {
	if e.Angry == "" {
		return DefaultAngryEmoji
	}
	return e.Angry
}

// FileEmoji returns the file emoji
func (e *Emoji) FileEmoji() string {
	if e.File == "" {
		return DefaultFileEmoji
	}
	return e.File
}

// FolderEmoji returns the folder emoji
func (e *Emoji) FolderEmoji() string {
	if e.Folder == "" {
		return DefaultFolderEmoji
	}
	return e.Folder
}

// CheckMarkEmoji returns the check mark emoji
func (e *Emoji) CheckMarkEmoji() string {
	if e.CheckMark == "" {
		return DefaultCheckMarkEmoji
	}
	return e.CheckMark
}

// LogStreamCancelledEmoji returns the log stream cancelled emoji
func (e *Emoji) LogStreamCancelledEmoji() string {
	if e.LogStreamCancelled == "" {
		return DefaultLogStreamCancelledEmoji
	}
	return e.LogStreamCancelled
}

// NewVersionEmoji returns the new version emoji
func (e *Emoji) NewVersionEmoji() string {
	if e.NewVersion == "" {
		return DefaultNewVersionEmoji
	}
	return e.NewVersion
}

// --- Getters for XRay emojis ---

// NamespacesEmoji returns the XRay namespaces emoji
func (e *XRayEmoji) NamespacesEmoji() string {
	if e.Namespaces == "" {
		return DefaultXRayNamespacesEmoji
	}
	return e.Namespaces
}

// DefaultGvrEmoji returns the XRay default GVR emoji
func (e *XRayEmoji) DefaultGvrEmoji() string {
	if e.DefaultGvr == "" {
		return DefaultXRayDefaultGvrEmoji
	}
	return e.DefaultGvr
}

// NodesEmoji returns the XRay nodes emoji
func (e *XRayEmoji) NodesEmoji() string {
	if e.Nodes == "" {
		return DefaultXRayNodesEmoji
	}
	return e.Nodes
}

// PodsEmoji returns the XRay pods emoji
func (e *XRayEmoji) PodsEmoji() string {
	if e.Pods == "" {
		return DefaultXRayPodsEmoji
	}
	return e.Pods
}

// ServicesEmoji returns the XRay services emoji
func (e *XRayEmoji) ServicesEmoji() string {
	if e.Services == "" {
		return DefaultXRayServicesEmoji
	}
	return e.Services
}

// ServiceAccountsEmoji returns the XRay service accounts emoji
func (e *XRayEmoji) ServiceAccountsEmoji() string {
	if e.ServiceAccounts == "" {
		return DefaultXRayServiceAccountsEmoji
	}
	return e.ServiceAccounts
}

// PersistentVolumesEmoji returns the XRay persistent volumes emoji
func (e *XRayEmoji) PersistentVolumesEmoji() string {
	if e.PersistentVolumes == "" {
		return DefaultXRayPersistentVolumesEmoji
	}
	return e.PersistentVolumes
}

// PersistentVolumeClaimsEmoji returns the XRay persistent volume claims emoji
func (e *XRayEmoji) PersistentVolumeClaimsEmoji() string {
	if e.PersistentVolumeClaims == "" {
		return DefaultXRayPersistentVolumeClaimsEmoji
	}
	return e.PersistentVolumeClaims
}

// SecretsEmoji returns the XRay secrets emoji
func (e *XRayEmoji) SecretsEmoji() string {
	if e.Secrets == "" {
		return DefaultXRaySecretsEmoji
	}
	return e.Secrets
}

// HorizontalPodAutoscalersEmoji returns the XRay horizontal pod autoscalers emoji
func (e *XRayEmoji) HorizontalPodAutoscalersEmoji() string {
	if e.HorizontalPodAutoscalers == "" {
		return DefaultXRayHorizontalPodAutoscalersEmoji
	}
	return e.HorizontalPodAutoscalers
}

// ConfigMapsEmoji returns the XRay config maps emoji
func (e *XRayEmoji) ConfigMapsEmoji() string {
	if e.ConfigMaps == "" {
		return DefaultXRayConfigMapsEmoji
	}
	return e.ConfigMaps
}

// DeploymentsEmoji returns the XRay deployments emoji
func (e *XRayEmoji) DeploymentsEmoji() string {
	if e.Deployments == "" {
		return DefaultXRayDeploymentsEmoji
	}
	return e.Deployments
}

// StatefulSetsEmoji returns the XRay stateful sets emoji
func (e *XRayEmoji) StatefulSetsEmoji() string {
	if e.StatefulSets == "" {
		return DefaultXRayStatefulSetsEmoji
	}
	return e.StatefulSets
}

// DaemonSetsEmoji returns the XRay daemon sets emoji
func (e *XRayEmoji) DaemonSetsEmoji() string {
	if e.DaemonSets == "" {
		return DefaultXRayDaemonSetsEmoji
	}
	return e.DaemonSets
}

// ReplicaSetsEmoji returns the XRay replica sets emoji
func (e *XRayEmoji) ReplicaSetsEmoji() string {
	if e.ReplicaSets == "" {
		return DefaultXRayReplicaSetsEmoji
	}
	return e.ReplicaSets
}

// ClusterRolesEmoji returns the XRay cluster roles emoji
func (e *XRayEmoji) ClusterRolesEmoji() string {
	if e.ClusterRoles == "" {
		return DefaultXRayClusterRolesEmoji
	}
	return e.ClusterRoles
}

// RolesEmoji returns the XRay roles emoji
func (e *XRayEmoji) RolesEmoji() string {
	if e.Roles == "" {
		return DefaultXRayRolesEmoji
	}
	return e.Roles
}

// NetworkPoliciesEmoji returns the XRay network policies emoji
func (e *XRayEmoji) NetworkPoliciesEmoji() string {
	if e.NetworkPolices == "" {
		return DefaultXRayNetworkPolicesEmoji
	}
	return e.NetworkPolices
}

// PodDisruptionBudgetsEmoji returns the XRay pod disruption budgets emoji
func (e *XRayEmoji) PodDisruptionBudgetsEmoji() string {
	if e.PodDisruptionBudgets == "" {
		return DefaultXRayPodDisruptionBudgetsEmoji
	}
	return e.PodDisruptionBudgets
}

// PodSecurityPoliciesEmoji returns the XRay pod security policies emoji
func (e *XRayEmoji) PodSecurityPoliciesEmoji() string {
	if e.PodSecurityPolicies == "" {
		return DefaultXRayPodSecurityPoliciesEmoji
	}
	return e.PodSecurityPolicies
}

// ContainersEmoji returns the XRay containers emoji
func (e *XRayEmoji) ContainersEmoji() string {
	if e.Containers == "" {
		return DefaultXRayContainersEmoji
	}
	return e.Containers
}

// ReportEmoji returns the XRay report emoji
func (e *XRayEmoji) ReportEmoji() string {
	if e.Report == "" {
		return DefaultXRayReportEmoji
	}
	return e.Report
}

// Issue0Emoji returns the XRay issue 0 emoji
func (e *XRayEmoji) Issue0Emoji() string {
	if e.Issue0 == "" {
		return DefaultXRayIssue0Emoji
	}
	return e.Issue0
}

// Issue1Emoji returns the XRay issue 1 emoji
func (e *XRayEmoji) Issue1Emoji() string {
	if e.Issue1 == "" {
		return DefaultXRayIssue1Emoji
	}
	return e.Issue1
}

// Issue2Emoji returns the XRay issue 2 emoji
func (e *XRayEmoji) Issue2Emoji() string {
	if e.Issue2 == "" {
		return DefaultXRayIssue2Emoji
	}
	return e.Issue2
}

// Issue3Emoji returns the XRay issue 3 emoji
func (e *XRayEmoji) Issue3Emoji() string {
	if e.Issue3 == "" {
		return DefaultXRayIssue3Emoji
	}
	return e.Issue3
}

const (
	DefaultStartUpEmoji            = "🐶"
	DefaultCommandLineEmoji        = "🐶"
	DefaultFilterLineEmoji         = "🐩"
	DefaultHappyEmoji              = "😎"
	DefaultWarnEmoji               = "😗"
	DefaultAngryEmoji              = "😡"
	DefaultFileEmoji               = "🦄"
	DefaultFolderEmoji             = "📁"
	DefaultCheckMarkEmoji          = "✅"
	DefaultLogStreamCancelledEmoji = "🏁"
	DefaultNewVersionEmoji         = "⚡️"

	DefaultXRayNamespacesEmoji               = "🗂"
	DefaultXRayDefaultGvrEmoji               = "📎"
	DefaultXRayNodesEmoji                    = "🖥"
	DefaultXRayPodsEmoji                     = "🚛"
	DefaultXRayServicesEmoji                 = "💁‍♀️"
	DefaultXRayServiceAccountsEmoji          = "💳"
	DefaultXRayPersistentVolumesEmoji        = "📚"
	DefaultXRayPersistentVolumeClaimsEmoji   = "🎟"
	DefaultXRaySecretsEmoji                  = "🔒"
	DefaultXRayHorizontalPodAutoscalersEmoji = "♎️"
	DefaultXRayConfigMapsEmoji               = "🗺"
	DefaultXRayDeploymentsEmoji              = "🪂"
	DefaultXRayStatefulSetsEmoji             = "🎎"
	DefaultXRayDaemonSetsEmoji               = "😈"
	DefaultXRayReplicaSetsEmoji              = "👯‍♂️"
	DefaultXRayClusterRolesEmoji             = "👩‍"
	DefaultXRayRolesEmoji                    = "👨🏻‍"
	DefaultXRayNetworkPolicesEmoji           = "📕"
	DefaultXRayPodDisruptionBudgetsEmoji     = "🏷"
	DefaultXRayPodSecurityPoliciesEmoji      = "👮‍♂️"
	DefaultXRayContainersEmoji               = "🐳"
	DefaultXRayReportEmoji                   = "🧼"
	DefaultXRayIssue0Emoji                   = "👍"
	DefaultXRayIssue1Emoji                   = "🔊"
	DefaultXRayIssue2Emoji                   = "☣️"
	DefaultXRayIssue3Emoji                   = "🧨"
)
