// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/render/helm"
	"github.com/derailed/k9s/internal/xray"
)

// Registry tracks resources metadata.
// BOZO!! Break up deps and merge into single registrar.
var Registry = map[string]ResourceMeta{
	// Custom...
	client.WkGVR.String(): {
		DAO:      new(dao.Workload),
		Renderer: new(render.Workload),
	},
	client.RefGVR.String(): {
		DAO:      new(dao.Reference),
		Renderer: new(render.Reference),
	},
	client.DirGVR.String(): {
		DAO:      new(dao.Dir),
		Renderer: new(render.Dir),
	},
	client.PuGVR.String(): {
		DAO: new(dao.Pulse),
	},
	client.HmGVR.String(): {
		DAO:      new(dao.HelmChart),
		Renderer: new(helm.Chart),
	},
	client.HmhGVR.String(): {
		DAO:      new(dao.HelmHistory),
		Renderer: new(helm.History),
	},
	client.CoGVR.String(): {
		DAO:          new(dao.Container),
		Renderer:     new(render.Container),
		TreeRenderer: new(xray.Container),
	},
	client.ScnGVR.String(): {
		DAO:      new(dao.ImageScan),
		Renderer: new(render.ImageScan),
	},
	client.CtGVR.String(): {
		DAO:      new(dao.Context),
		Renderer: new(render.Context),
	},
	client.SdGVR.String(): {
		DAO:      new(dao.ScreenDump),
		Renderer: new(render.ScreenDump),
	},
	client.RbacGVR.String(): {
		DAO:      new(dao.Rbac),
		Renderer: new(render.Rbac),
	},
	client.PolGVR.String(): {
		DAO:      new(dao.Policy),
		Renderer: new(render.Policy),
	},
	client.UsrGVR.String(): {
		DAO:      new(dao.Subject),
		Renderer: new(render.Subject),
	},
	client.GrpGVR.String(): {
		DAO:      new(dao.Subject),
		Renderer: new(render.Subject),
	},
	client.PfGVR.String(): {
		DAO:      new(dao.PortForward),
		Renderer: new(render.PortForward),
	},
	client.BeGVR.String(): {
		DAO:      new(dao.Benchmark),
		Renderer: new(render.Benchmark),
	},
	client.AliGVR.String(): {
		DAO:      new(dao.Alias),
		Renderer: new(render.Alias),
	},

	// Core...
	client.EpGVR.String(): {
		Renderer: new(render.Endpoints),
	},
	client.PodGVR.String(): {
		DAO:          new(dao.Pod),
		Renderer:     render.NewPod(),
		TreeRenderer: new(xray.Pod),
	},
	client.NsGVR.String(): {
		DAO:      new(dao.Namespace),
		Renderer: new(render.Namespace),
	},
	client.SecGVR.String(): {
		DAO:      new(dao.Secret),
		Renderer: new(render.Secret),
	},
	client.CmGVR.String(): {
		DAO:      new(dao.ConfigMap),
		Renderer: new(render.ConfigMap),
	},
	client.NodeGVR.String(): {
		DAO:      new(dao.Node),
		Renderer: new(render.Node),
	},
	client.SvcGVR.String(): {
		DAO:          new(dao.Service),
		Renderer:     new(render.Service),
		TreeRenderer: new(xray.Service),
	},
	client.SaGVR.String(): {
		Renderer: new(render.ServiceAccount),
	},
	client.PvGVR.String(): {
		Renderer: new(render.PersistentVolume),
	},
	client.PvcGVR.String(): {
		Renderer: new(render.PersistentVolumeClaim),
	},

	// Apps...
	client.DpGVR.String(): {
		DAO:          new(dao.Deployment),
		Renderer:     new(render.Deployment),
		TreeRenderer: new(xray.Deployment),
	},
	client.RsGVR.String(): {
		Renderer:     new(render.ReplicaSet),
		TreeRenderer: new(xray.ReplicaSet),
	},
	client.StsGVR.String(): {
		DAO:          new(dao.StatefulSet),
		Renderer:     new(render.StatefulSet),
		TreeRenderer: new(xray.StatefulSet),
	},
	client.DsGVR.String(): {
		DAO:          new(dao.DaemonSet),
		Renderer:     new(render.DaemonSet),
		TreeRenderer: new(xray.DaemonSet),
	},

	// Extensions...
	client.NpGVR.String(): {
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	client.CjGVR.String(): {
		DAO:      new(dao.CronJob),
		Renderer: new(render.CronJob),
	},
	client.JobGVR.String(): {
		DAO:      new(dao.Job),
		Renderer: new(render.Job),
	},

	// CRDs...
	client.CrdGVR.String(): {
		DAO:      new(dao.CustomResourceDefinition),
		Renderer: new(render.CustomResourceDefinition),
	},

	// Storage...
	client.ScGVR.String(): {
		Renderer: &render.StorageClass{},
	},

	// Policy...
	client.PdbGVR.String(): {
		Renderer: &render.PodDisruptionBudget{},
	},

	// RBAC...
	client.CrGVR.String(): {
		DAO:      new(dao.Rbac),
		Renderer: new(render.ClusterRole),
	},
	client.CrbGVR.String(): {
		Renderer: new(render.ClusterRoleBinding),
	},
	client.RoGVR.String(): {
		Renderer: new(render.Role),
	},
	client.RobGVR.String(): {
		Renderer: new(render.RoleBinding),
	},
}
