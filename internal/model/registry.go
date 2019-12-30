package model

import (
	"github.com/derailed/k9s/internal/render"
)

// Registry tracks resources metadata.
// BOZO!! Break up deps and merge into single registrar
var Registry = map[string]ResourceMeta{
	// Custom...
	"containers": {
		Model:    &Container{},
		Renderer: &render.Container{},
	},
	"contexts": {
		Model:    &Context{},
		Renderer: &render.Context{},
	},
	"screendumps": {
		Model:    &ScreenDump{},
		Renderer: &render.ScreenDump{},
	},
	"rbac": {
		Model:    &Rbac{},
		Renderer: &render.Rbac{},
	},
	"policy": {
		Model:    &Policy{},
		Renderer: &render.Policy{},
	},
	"users": {
		Model:    &Subject{},
		Renderer: &render.Subject{},
	},
	"groups": {
		Model:    &Subject{},
		Renderer: &render.Subject{},
	},
	"portforwards": {
		Model:    &PortForward{},
		Renderer: &render.PortForward{},
	},
	"benchmarks": {
		Model:    &Benchmark{},
		Renderer: &render.Benchmark{},
	},
	"aliases": {
		Model:    &Alias{},
		Renderer: &render.Alias{},
	},

	// Core...
	"v1/configmaps": {
		Renderer: &render.ConfigMap{},
	},
	"v1/endpoints": {
		Renderer: &render.Endpoints{},
	},
	"v1/events": {
		Renderer: &render.Event{},
	},
	"v1/pods": {
		Model:    &Pod{},
		Renderer: &render.Pod{},
	},
	"v1/namespaces": {
		Renderer: &render.Namespace{},
	},
	"v1/nodes": {
		Model:    &Node{},
		Renderer: &render.Node{},
	},
	"v1/secrets": {
		Renderer: &render.Secret{},
	},
	"v1/services": {
		Renderer: &render.Service{},
	},
	"v1/serviceaccounts": {
		Renderer: &render.ServiceAccount{},
	},

	// Apps...
	"apps/v1/deployments": {
		Renderer: &render.Deployment{},
	},
	"apps/v1/replicasets": {
		Renderer: &render.ReplicaSet{},
	},
	"apps/v1/statefulsets": {
		Renderer: &render.StatefulSet{},
	},
	"apps/v1/daemonsets": {
		Renderer: &render.DaemonSet{},
	},

	// Extensions...
	"extensions/v1beta1/daemonsets": {
		Renderer: &render.DaemonSet{},
	},
	"extensions/v1beta1/ingresses": {
		Renderer: &render.Ingress{},
	},
	"extensions/v1beta1/networkpolicies": {
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	"batch/v1beta1/cronjobs": {
		Renderer: &render.CronJob{},
	},
	"batch/v1/jobs": {
		Model:    &Job{},
		Renderer: &render.Job{},
	},

	// Autoscaling...
	"autoscaling/v1/horizontalpodautoscalers": {
		Model:    &HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},
	"autoscaling/v2beta1/horizontalpodautoscalers": {
		Model:    &HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},
	"autoscaling/v2beta2/horizontalpodautoscalers": {
		Model:    &HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},

	// CRDs...
	"apiextensions.k8s.io/v1/customresourcedefinitions": {
		Model:    &CustomResourceDefinition{},
		Renderer: &render.CustomResourceDefinition{},
	},
	"apiextensions.k8s.io/v1beta1/customresourcedefinitions": {
		Model:    &CustomResourceDefinition{},
		Renderer: &render.CustomResourceDefinition{},
	},

	// Storage...
	"storage.k8s.io/v1/storageclasses": {
		Renderer: &render.StorageClass{},
	},

	// Policy...
	"policy/v1beta1/poddisruptionbudgets": {
		Renderer: &render.PodDisruptionBudget{},
	},

	// RBAC...
	"rbac.authorization.k8s.io/v1/clusterroles": {
		Model:    &Rbac{},
		Renderer: &render.ClusterRole{},
	},
	"rbac.authorization.k8s.io/v1/clusterrolebindings": {
		Renderer: &render.ClusterRoleBinding{},
	},
	"rbac.authorization.k8s.io/v1/roles": {
		Renderer: &render.Role{},
	},
	"rbac.authorization.k8s.io/v1/rolebindings": {
		Renderer: &render.RoleBinding{},
	},
}
