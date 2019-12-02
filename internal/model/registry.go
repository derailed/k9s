package model

import (
	"github.com/derailed/k9s/internal/render"
)

// BOZO!! Break up deps and merge into single registrar
var Registry = map[string]ResourceMeta{
	"containers": ResourceMeta{
		Model:    &Container{},
		Renderer: &render.Container{},
	},
	"contexts": ResourceMeta{
		Model:    &Context{},
		Renderer: &render.Context{},
	},
	"screendumps": ResourceMeta{
		Model:    &ScreenDump{},
		Renderer: &render.ScreenDump{},
	},

	"v1/pods": ResourceMeta{
		Model:    &Pod{},
		Renderer: &render.Pod{},
	},
	"v1/nodes": ResourceMeta{
		Model:    &Node{},
		Renderer: &render.Node{},
	},
	"v1/namespaces": ResourceMeta{
		Renderer: &render.Namespace{},
	},

	"apps/v1/deployments": ResourceMeta{
		Renderer: &render.Deployment{},
	},
	"apps/v1/replicasets": ResourceMeta{
		Renderer: &render.ReplicaSet{},
	},
	"apps/v1/statefulsets": ResourceMeta{
		Renderer: &render.StatefulSet{},
	},
	"apps/v1/daemonsets": ResourceMeta{
		Renderer: &render.DaemonSet{},
	},
	"extensions/v1beta1/daemonsets": ResourceMeta{
		Renderer: &render.DaemonSet{},
	},

	// "v1/services": ResourceMeta{
	// 	Renderer: &render.Service{},
	// },
	// "v1/configmaps": ResourceMeta{
	// 	Renderer: &render.ConfigMap{},
	// },
	// "v1/secrets": ResourceMeta{
	// 	Renderer: &render.ConfigMap{},
	// },
	// "batch/v1beta1/cronjobs": ResourceMeta{
	// 	Renderer: &render.CronJob{},
	// },
	// "batch/v1/jobs": ResourceMeta{
	// 	Renderer: &render.Job{},
	// },

	"apiextensions.k8s.io/v1beta1/customresourcedefinitions": ResourceMeta{
		Renderer: &render.CustomResourceDefinition{},
	},

	"rbac.authorization.k8s.io/v1/clusterroles": ResourceMeta{
		Renderer: &render.ClusterRole{},
	},
	"rbac.authorization.k8s.io/v1/clusterrolebindings": ResourceMeta{
		Renderer: &render.ClusterRoleBinding{},
	},
	"rbac.authorization.k8s.io/v1/roles": ResourceMeta{
		Renderer: &render.Role{},
	},
	"rbac.authorization.k8s.io/v1/rolebindings": ResourceMeta{
		Renderer: &render.RoleBinding{},
	},
}
