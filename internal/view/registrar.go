// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"github.com/derailed/k9s/internal/client"
)

func loadCustomViewers() MetaViewers {
	m := make(MetaViewers, 30)
	coreViewers(m)
	miscViewers(m)
	appsViewers(m)
	rbacViewers(m)
	batchViewers(m)
	crdViewers(m)
	helmViewers(m)

	return m
}

func helmViewers(vv MetaViewers) {
	vv[client.NewGVR("helm")] = MetaViewer{
		viewerFn: NewHelmChart,
	}
}

func coreViewers(vv MetaViewers) {
	vv[client.NewGVR("v1/namespaces")] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv[client.NewGVR("v1/events")] = MetaViewer{
		viewerFn: NewEvent,
	}
	vv[client.NewGVR("v1/pods")] = MetaViewer{
		viewerFn: NewPod,
	}
	vv[client.NewGVR("v1/services")] = MetaViewer{
		viewerFn: NewService,
	}
	vv[client.NewGVR("v1/nodes")] = MetaViewer{
		viewerFn: NewNode,
	}
	vv[client.NewGVR("v1/secrets")] = MetaViewer{
		viewerFn: NewSecret,
	}
	vv[client.NewGVR("scheduling.k8s.io/v1/priorityclasses")] = MetaViewer{
		viewerFn: NewPriorityClass,
	}
	vv[client.NewGVR("v1/configmaps")] = MetaViewer{
		viewerFn: NewConfigMap,
	}
	vv[client.NewGVR("v1/serviceaccounts")] = MetaViewer{
		viewerFn: NewServiceAccount,
	}
	vv[client.NewGVR("v1/persistentvolumeclaims")] = MetaViewer{
		viewerFn: NewPersistentVolumeClaim,
	}
}

func miscViewers(vv MetaViewers) {
	vv[client.NewGVR("workloads")] = MetaViewer{
		viewerFn: NewWorkload,
	}
	vv[client.NewGVR("contexts")] = MetaViewer{
		viewerFn: NewContext,
	}
	vv[client.NewGVR("containers")] = MetaViewer{
		viewerFn: NewContainer,
	}
	vv[client.NewGVR("scans")] = MetaViewer{
		viewerFn: NewImageScan,
	}
	vv[client.NewGVR("portforwards")] = MetaViewer{
		viewerFn: NewPortForward,
	}
	vv[client.NewGVR("screendumps")] = MetaViewer{
		viewerFn: NewScreenDump,
	}
	vv[client.NewGVR("benchmarks")] = MetaViewer{
		viewerFn: NewBenchmark,
	}
	vv[client.NewGVR("aliases")] = MetaViewer{
		viewerFn: NewAlias,
	}
	vv[client.NewGVR("references")] = MetaViewer{
		viewerFn: NewReference,
	}
	vv[client.NewGVR("pulses")] = MetaViewer{
		viewerFn: NewPulse,
	}
	// !!BOZO!! Popeye
	// vv[client.NewGVR("popeye")] = MetaViewer{
	// 	viewerFn: NewPopeye,
	// }
	vv[client.NewGVR("sanitizer")] = MetaViewer{
		viewerFn: NewSanitizer,
	}
}

func appsViewers(vv MetaViewers) {
	vv[client.NewGVR("apps/v1/deployments")] = MetaViewer{
		viewerFn: NewDeploy,
	}
	vv[client.NewGVR("apps/v1/replicasets")] = MetaViewer{
		viewerFn: NewReplicaSet,
	}
	vv[client.NewGVR("apps/v1/statefulsets")] = MetaViewer{
		viewerFn: NewStatefulSet,
	}
	vv[client.NewGVR("apps/v1/daemonsets")] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
	vv[client.NewGVR("apps/v1/daemonsets")] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
}

func rbacViewers(vv MetaViewers) {
	vv[client.NewGVR("rbac")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("users")] = MetaViewer{
		viewerFn: NewUser,
	}
	vv[client.NewGVR("groups")] = MetaViewer{
		viewerFn: NewGroup,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/clusterroles")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/roles")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/clusterrolebindings")] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.NewGVR("rbac.authorization.k8s.io/v1/rolebindings")] = MetaViewer{
		enterFn: showRules,
	}
}

func batchViewers(vv MetaViewers) {
	vv[client.NewGVR("batch/v1beta1/cronjobs")] = MetaViewer{
		viewerFn: NewCronJob,
	}
	vv[client.NewGVR("batch/v1/cronjobs")] = MetaViewer{
		viewerFn: NewCronJob,
	}
	vv[client.NewGVR("batch/v1/jobs")] = MetaViewer{
		viewerFn: NewJob,
	}
}

func crdViewers(vv MetaViewers) {
	vv[client.NewGVR("apiextensions.k8s.io/v1/customresourcedefinitions")] = MetaViewer{
		viewerFn: NewCRD,
	}
}
