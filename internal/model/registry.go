package model

import (
	"github.com/derailed/k9s/internal/render"
)

// BOZO!! Break up deps and merge into single registrar
var Registry = map[string]ResourceMeta{
	// Custom...
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
	"rbac": ResourceMeta{
		Model:    &Rbac{},
		Renderer: &render.Rbac{},
	},
	"portforwards": ResourceMeta{
		Model:    &PortForward{},
		Renderer: &render.PortForward{},
	},
	"benchmarks": ResourceMeta{
		Model:    &Benchmark{},
		Renderer: &render.Benchmark{},
	},
	"aliases": ResourceMeta{
		Model:    &Alias{},
		Renderer: &render.Alias{},
	},

	// Core...
	"v1/configmaps": ResourceMeta{
		Renderer: &render.ConfigMap{},
	},
	"v1/endpoints": ResourceMeta{
		Renderer: &render.Endpoints{},
	},
	"v1/events": ResourceMeta{
		Renderer: &render.Event{},
	},
	"v1/pods": ResourceMeta{
		Model:    &Pod{},
		Renderer: &render.Pod{},
	},
	"v1/namespaces": ResourceMeta{
		Renderer: &render.Namespace{},
	},
	"v1/nodes": ResourceMeta{
		Model:    &Node{},
		Renderer: &render.Node{},
	},
	"v1/secrets": ResourceMeta{
		Renderer: &render.Secret{},
	},
	"v1/services": ResourceMeta{
		Renderer: &render.Service{},
	},
	"v1/serviceaccounts": ResourceMeta{
		Renderer: &render.ServiceAccount{},
	},

	// Apps...
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

	// Extensions...
	"extensions/v1beta1/daemonsets": ResourceMeta{
		Renderer: &render.DaemonSet{},
	},
	"extensions/v1beta1/ingresses": ResourceMeta{
		Renderer: &render.Ingress{},
	},
	"extensions/v1beta1/networkpolicies": ResourceMeta{
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	"batch/v1beta1/cronjobs": ResourceMeta{
		Renderer: &render.CronJob{},
	},
	"batch/v1/jobs": ResourceMeta{
		Model:    &Job{},
		Renderer: &render.Job{},
	},

	// Autoscaling...
	"autoscaling/v1/horizontalpodautoscalers": ResourceMeta{
		Renderer: &render.HorizontalPodAutoscaler{},
	},

	// CRDs...
	"apiextensions.k8s.io/v1beta1/customresourcedefinitions": ResourceMeta{
		Renderer: &render.CustomResourceDefinition{},
	},

	// Storage...
	"storage.k8s.io/v1/storageclasses": ResourceMeta{
		Renderer: &render.StorageClass{},
	},

	// Policy...
	"policy/v1beta1/poddisruptionbudgets": ResourceMeta{
		Renderer: &render.PodDisruptionBudget{},
	},

	// RBAC...
	"rbac.authorization.k8s.io/v1/clusterroles": ResourceMeta{
		Model:    &Rbac{},
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
