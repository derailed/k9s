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
	vv[client.HmGVR] = MetaViewer{
		viewerFn: NewHelmChart,
	}
}

func coreViewers(vv MetaViewers) {
	vv[client.NsGVR] = MetaViewer{
		viewerFn: NewNamespace,
	}
	vv[client.EvGVR] = MetaViewer{
		viewerFn: NewEvent,
	}
	vv[client.PodGVR] = MetaViewer{
		viewerFn: NewPod,
	}
	vv[client.SvcGVR] = MetaViewer{
		viewerFn: NewService,
	}
	vv[client.NodeGVR] = MetaViewer{
		viewerFn: NewNode,
	}
	vv[client.SecGVR] = MetaViewer{
		viewerFn: NewSecret,
	}
	vv[client.PcGVR] = MetaViewer{
		viewerFn: NewPriorityClass,
	}
	vv[client.CmGVR] = MetaViewer{
		viewerFn: NewConfigMap,
	}
	vv[client.SaGVR] = MetaViewer{
		viewerFn: NewServiceAccount,
	}
	vv[client.PvcGVR] = MetaViewer{
		viewerFn: NewPersistentVolumeClaim,
	}
}

func miscViewers(vv MetaViewers) {
	vv[client.WkGVR] = MetaViewer{
		viewerFn: NewWorkload,
	}
	vv[client.CtGVR] = MetaViewer{
		viewerFn: NewContext,
	}
	vv[client.CoGVR] = MetaViewer{
		viewerFn: NewContainer,
	}
	vv[client.ScnGVR] = MetaViewer{
		viewerFn: NewImageScan,
	}
	vv[client.PfGVR] = MetaViewer{
		viewerFn: NewPortForward,
	}
	vv[client.SdGVR] = MetaViewer{
		viewerFn: NewScreenDump,
	}
	vv[client.BeGVR] = MetaViewer{
		viewerFn: NewBenchmark,
	}
	vv[client.AliGVR] = MetaViewer{
		viewerFn: NewAlias,
	}
	vv[client.RefGVR] = MetaViewer{
		viewerFn: NewReference,
	}
	vv[client.PuGVR] = MetaViewer{
		viewerFn: NewPulse,
	}
}

func appsViewers(vv MetaViewers) {
	vv[client.DpGVR] = MetaViewer{
		viewerFn: NewDeploy,
	}
	vv[client.RsGVR] = MetaViewer{
		viewerFn: NewReplicaSet,
	}
	vv[client.StsGVR] = MetaViewer{
		viewerFn: NewStatefulSet,
	}
	vv[client.DsGVR] = MetaViewer{
		viewerFn: NewDaemonSet,
	}
}

func rbacViewers(vv MetaViewers) {
	vv[client.RbacGVR] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.UsrGVR] = MetaViewer{
		viewerFn: NewUser,
	}
	vv[client.GrpGVR] = MetaViewer{
		viewerFn: NewGroup,
	}
	vv[client.CrGVR] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.CrbGVR] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.RoGVR] = MetaViewer{
		enterFn: showRules,
	}
	vv[client.RobGVR] = MetaViewer{
		enterFn: showRules,
	}
}

func batchViewers(vv MetaViewers) {
	vv[client.CjGVR] = MetaViewer{
		viewerFn: NewCronJob,
	}
	vv[client.JobGVR] = MetaViewer{
		viewerFn: NewJob,
	}
}

func crdViewers(vv MetaViewers) {
	vv[client.CrdGVR] = MetaViewer{
		viewerFn: NewCRD,
	}
}
