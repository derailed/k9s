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
var Registry = map[*client.GVR]ResourceMeta{
	// Custom...
	client.WkGVR: {
		DAO:      new(dao.Workload),
		Renderer: new(render.Workload),
	},
	client.RefGVR: {
		DAO:      new(dao.Reference),
		Renderer: new(render.Reference),
	},
	client.DirGVR: {
		DAO:      new(dao.Dir),
		Renderer: new(render.Dir),
	},
	client.PuGVR: {
		DAO: new(dao.Pulse),
	},
	client.HmGVR: {
		DAO:      new(dao.HelmChart),
		Renderer: new(helm.Chart),
	},
	client.HmhGVR: {
		DAO:      new(dao.HelmHistory),
		Renderer: new(helm.History),
	},
	client.CoGVR: {
		DAO:          new(dao.Container),
		Renderer:     new(render.Container),
		TreeRenderer: new(xray.Container),
	},
	client.ScnGVR: {
		DAO:      new(dao.ImageScan),
		Renderer: new(render.ImageScan),
	},
	client.CtGVR: {
		DAO:      new(dao.Context),
		Renderer: new(render.Context),
	},
	client.SdGVR: {
		DAO:      new(dao.ScreenDump),
		Renderer: new(render.ScreenDump),
	},
	client.RbacGVR: {
		DAO:      new(dao.Rbac),
		Renderer: new(render.Rbac),
	},
	client.PolGVR: {
		DAO:      new(dao.Policy),
		Renderer: new(render.Policy),
	},
	client.UsrGVR: {
		DAO:      new(dao.Subject),
		Renderer: new(render.Subject),
	},
	client.GrpGVR: {
		DAO:      new(dao.Subject),
		Renderer: new(render.Subject),
	},
	client.PfGVR: {
		DAO:      new(dao.PortForward),
		Renderer: new(render.PortForward),
	},
	client.BeGVR: {
		DAO:      new(dao.Benchmark),
		Renderer: new(render.Benchmark),
	},
	client.AliGVR: {
		DAO:      new(dao.Alias),
		Renderer: new(render.Alias),
	},

	// Core...
	client.EpGVR: {
		Renderer: new(render.Endpoints),
	},
	client.PodGVR: {
		DAO:          new(dao.Pod),
		Renderer:     render.NewPod(),
		TreeRenderer: new(xray.Pod),
	},
	client.NsGVR: {
		DAO:      new(dao.Namespace),
		Renderer: new(render.Namespace),
	},
	client.SecGVR: {
		DAO:      new(dao.Secret),
		Renderer: new(render.Secret),
	},
	client.CmGVR: {
		DAO:      new(dao.ConfigMap),
		Renderer: new(render.ConfigMap),
	},
	client.NodeGVR: {
		DAO:      new(dao.Node),
		Renderer: new(render.Node),
	},
	client.SvcGVR: {
		DAO:          new(dao.Service),
		Renderer:     new(render.Service),
		TreeRenderer: new(xray.Service),
	},
	client.SaGVR: {
		Renderer: new(render.ServiceAccount),
	},
	client.PvGVR: {
		Renderer: new(render.PersistentVolume),
	},
	client.PvcGVR: {
		Renderer: new(render.PersistentVolumeClaim),
	},
	client.EvGVR: {
		Renderer: new(render.Event),
	},

	// Apps...
	client.DpGVR: {
		DAO:          new(dao.Deployment),
		Renderer:     new(render.Deployment),
		TreeRenderer: new(xray.Deployment),
	},
	client.RsGVR: {
		Renderer:     new(render.ReplicaSet),
		TreeRenderer: new(xray.ReplicaSet),
	},
	client.StsGVR: {
		DAO:          new(dao.StatefulSet),
		Renderer:     new(render.StatefulSet),
		TreeRenderer: new(xray.StatefulSet),
	},
	client.DsGVR: {
		DAO:          new(dao.DaemonSet),
		Renderer:     new(render.DaemonSet),
		TreeRenderer: new(xray.DaemonSet),
	},

	// Extensions...
	client.NpGVR: {
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	client.CjGVR: {
		DAO:      new(dao.CronJob),
		Renderer: new(render.CronJob),
	},
	client.JobGVR: {
		DAO:      new(dao.Job),
		Renderer: new(render.Job),
	},

	// CRDs...
	client.CrdGVR: {
		DAO:      new(dao.CustomResourceDefinition),
		Renderer: new(render.CustomResourceDefinition),
	},

	// Storage...
	client.ScGVR: {
		Renderer: &render.StorageClass{},
	},

	// Policy...
	client.PdbGVR: {
		Renderer: &render.PodDisruptionBudget{},
	},

	// RBAC...
	client.CrGVR: {
		DAO:      new(dao.Rbac),
		Renderer: new(render.ClusterRole),
	},
	client.CrbGVR: {
		Renderer: new(render.ClusterRoleBinding),
	},
	client.RoGVR: {
		Renderer: new(render.Role),
	},
	client.RobGVR: {
		Renderer: new(render.RoleBinding),
	},
}
