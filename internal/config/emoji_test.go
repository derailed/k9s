// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package config_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/derailed/k9s/internal/config"
)

func TestEmojiDefaults(t *testing.T) {
	var e config.Emoji

	// Test main emoji getters with empty struct
	assert.Equal(t, config.DefaultStartUpEmoji, e.StartUpEmoji())
	assert.Equal(t, config.DefaultCommandLineEmoji, e.CommandLineEmoji())
	assert.Equal(t, config.DefaultFilterLineEmoji, e.FilterLineEmoji())
	assert.Equal(t, config.DefaultHappyEmoji, e.InfoStatusEmoji())
	assert.Equal(t, config.DefaultWarnEmoji, e.WarningStatusEmoji())
	assert.Equal(t, config.DefaultAngryEmoji, e.ErrorStatusEmoji())
	assert.Equal(t, config.DefaultFileEmoji, e.FileEmoji())
	assert.Equal(t, config.DefaultFolderEmoji, e.FolderEmoji())
	assert.Equal(t, config.DefaultCheckMarkEmoji, e.CheckMarkEmoji())
	assert.Equal(t, config.DefaultLogStreamCancelledEmoji, e.LogStreamCancelledEmoji())
	assert.Equal(t, config.DefaultNewVersionEmoji, e.NewVersionEmoji())
}

func TestEmojiCustom(t *testing.T) {
	e := config.Emoji{
		StartUp:            "🇩🇰",
		CommandLine:        "🇳🇴",
		FilterLine:         "🇸🇪",
		Happy:              "😊",
		Warn:               "😐",
		Angry:              "😠",
		File:               "📄",
		Folder:             "📂",
		CheckMark:          "✓",
		LogStreamCancelled: "✖",
		NewVersion:         "🚀",
	}

	// Test main emoji getters with custom values
	assert.Equal(t, "🇩🇰", e.StartUpEmoji())
	assert.Equal(t, "🇳🇴", e.CommandLineEmoji())
	assert.Equal(t, "🇸🇪", e.FilterLineEmoji())
	assert.Equal(t, "😊", e.InfoStatusEmoji())
	assert.Equal(t, "😐", e.WarningStatusEmoji())
	assert.Equal(t, "😠", e.ErrorStatusEmoji())
	assert.Equal(t, "📄", e.FileEmoji())
	assert.Equal(t, "📂", e.FolderEmoji())
	assert.Equal(t, "✓", e.CheckMarkEmoji())
	assert.Equal(t, "✖", e.LogStreamCancelledEmoji())
	assert.Equal(t, "🚀", e.NewVersionEmoji())
}

func TestXRayEmojiDefaults(t *testing.T) {
	e := config.XRayEmoji{}

	// Test XRay emoji getters with empty struct
	assert.Equal(t, config.DefaultXRayNamespacesEmoji, e.NamespacesEmoji())
	assert.Equal(t, config.DefaultXRayDefaultGvrEmoji, e.DefaultGvrEmoji())
	assert.Equal(t, config.DefaultXRayNodesEmoji, e.NodesEmoji())
	assert.Equal(t, config.DefaultXRayPodsEmoji, e.PodsEmoji())
	assert.Equal(t, config.DefaultXRayServicesEmoji, e.ServicesEmoji())
	assert.Equal(t, config.DefaultXRayServiceAccountsEmoji, e.ServiceAccountsEmoji())
	assert.Equal(t, config.DefaultXRayPersistentVolumesEmoji, e.PersistentVolumesEmoji())
	assert.Equal(t, config.DefaultXRayPersistentVolumeClaimsEmoji, e.PersistentVolumeClaimsEmoji())
	assert.Equal(t, config.DefaultXRaySecretsEmoji, e.SecretsEmoji())
	assert.Equal(t, config.DefaultXRayHorizontalPodAutoscalersEmoji, e.HorizontalPodAutoscalersEmoji())
	assert.Equal(t, config.DefaultXRayConfigMapsEmoji, e.ConfigMapsEmoji())
	assert.Equal(t, config.DefaultXRayDeploymentsEmoji, e.DeploymentsEmoji())
	assert.Equal(t, config.DefaultXRayStatefulSetsEmoji, e.StatefulSetsEmoji())
	assert.Equal(t, config.DefaultXRayDaemonSetsEmoji, e.DaemonSetsEmoji())
	assert.Equal(t, config.DefaultXRayReplicaSetsEmoji, e.ReplicaSetsEmoji())
	assert.Equal(t, config.DefaultXRayClusterRolesEmoji, e.ClusterRolesEmoji())
	assert.Equal(t, config.DefaultXRayRolesEmoji, e.RolesEmoji())
	assert.Equal(t, config.DefaultXRayNetworkPolicesEmoji, e.NetworkPoliciesEmoji())
	assert.Equal(t, config.DefaultXRayPodDisruptionBudgetsEmoji, e.PodDisruptionBudgetsEmoji())
	assert.Equal(t, config.DefaultXRayPodSecurityPoliciesEmoji, e.PodSecurityPoliciesEmoji())
	assert.Equal(t, config.DefaultXRayContainersEmoji, e.ContainersEmoji())
	assert.Equal(t, config.DefaultXRayReportEmoji, e.ReportEmoji())
	assert.Equal(t, config.DefaultXRayIssue0Emoji, e.Issue0Emoji())
	assert.Equal(t, config.DefaultXRayIssue1Emoji, e.Issue1Emoji())
	assert.Equal(t, config.DefaultXRayIssue2Emoji, e.Issue2Emoji())
	assert.Equal(t, config.DefaultXRayIssue3Emoji, e.Issue3Emoji())
}

func TestXRayEmojiCustom(t *testing.T) {
	e := config.XRayEmoji{
		Namespaces:               "📁",
		DefaultGvr:               "📌",
		Nodes:                    "🖲",
		Pods:                     "🚗",
		Services:                 "🔌",
		ServiceAccounts:          "🔑",
		PersistentVolumes:        "💾",
		PersistentVolumeClaims:   "📝",
		Secrets:                  "🔐",
		HorizontalPodAutoscalers: "⚖️",
		ConfigMaps:               "🗂",
		Deployments:              "🚀",
		StatefulSets:             "📊",
		DaemonSets:               "👹",
		ReplicaSets:              "👥",
		ClusterRoles:             "👑",
		Roles:                    "👤",
		NetworkPolices:           "🔥",
		PodDisruptionBudgets:     "🏷️",
		PodSecurityPolicies:      "🛡️",
		Containers:               "📦",
		Report:                   "📜",
		Issue0:                   "👍",
		Issue1:                   "⚠️",
		Issue2:                   "⛔",
		Issue3:                   "💣",
	}

	// Test XRay emoji getters with custom values
	assert.Equal(t, "📁", e.NamespacesEmoji())
	assert.Equal(t, "📌", e.DefaultGvrEmoji())
	assert.Equal(t, "🖲", e.NodesEmoji())
	assert.Equal(t, "🚗", e.PodsEmoji())
	assert.Equal(t, "🔌", e.ServicesEmoji())
	assert.Equal(t, "🔑", e.ServiceAccountsEmoji())
	assert.Equal(t, "💾", e.PersistentVolumesEmoji())
	assert.Equal(t, "📝", e.PersistentVolumeClaimsEmoji())
	assert.Equal(t, "🔐", e.SecretsEmoji())
	assert.Equal(t, "⚖️", e.HorizontalPodAutoscalersEmoji())
	assert.Equal(t, "🗂", e.ConfigMapsEmoji())
	assert.Equal(t, "🚀", e.DeploymentsEmoji())
	assert.Equal(t, "📊", e.StatefulSetsEmoji())
	assert.Equal(t, "👹", e.DaemonSetsEmoji())
	assert.Equal(t, "👥", e.ReplicaSetsEmoji())
	assert.Equal(t, "👑", e.ClusterRolesEmoji())
	assert.Equal(t, "👤", e.RolesEmoji())
	assert.Equal(t, "🔥", e.NetworkPoliciesEmoji())
	assert.Equal(t, "🏷️", e.PodDisruptionBudgetsEmoji())
	assert.Equal(t, "🛡️", e.PodSecurityPoliciesEmoji())
	assert.Equal(t, "📦", e.ContainersEmoji())
	assert.Equal(t, "📜", e.ReportEmoji())
	assert.Equal(t, "👍", e.Issue0Emoji())
	assert.Equal(t, "⚠️", e.Issue1Emoji())
	assert.Equal(t, "⛔", e.Issue2Emoji())
	assert.Equal(t, "💣", e.Issue3Emoji())
}
