package model

import (
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
)

// Registry tracks resources metadata.
// BOZO!! Break up deps and merge into single registrar
var Registry = map[string]ResourceMeta{
	// Custom...
	"charts": {
		DAO:      &dao.Chart{},
		Renderer: &render.Chart{},
	},
	"containers": {
		DAO:      &dao.Container{},
		Renderer: &render.Container{},
	},
	"contexts": {
		DAO:      &dao.Context{},
		Renderer: &render.Context{},
	},
	"screendumps": {
		DAO:      &dao.ScreenDump{},
		Renderer: &render.ScreenDump{},
	},
	"rbac": {
		DAO:      &dao.Rbac{},
		Renderer: &render.Rbac{},
	},
	"policy": {
		DAO:      &dao.Policy{},
		Renderer: &render.Policy{},
	},
	"users": {
		DAO:      &dao.Subject{},
		Renderer: &render.Subject{},
	},
	"groups": {
		DAO:      &dao.Subject{},
		Renderer: &render.Subject{},
	},
	"portforwards": {
		DAO:      &dao.PortForward{},
		Renderer: &render.PortForward{},
	},
	"benchmarks": {
		DAO:      &dao.Benchmark{},
		Renderer: &render.Benchmark{},
	},
	"aliases": {
		DAO:      &dao.Alias{},
		Renderer: &render.Alias{},
	},

	// Core...
	"v1/endpoints": {
		Renderer: &render.Endpoints{},
	},
	"v1/events": {
		Renderer: &render.Event{},
	},
	"v1/pods": {
		DAO:      &dao.Pod{},
		Renderer: &render.Pod{},
	},
	"v1/namespaces": {
		Renderer: &render.Namespace{},
	},
	"v1/nodes": {
		DAO:      &dao.Node{},
		Renderer: &render.Node{},
	},
	"v1/services": {
		Renderer: &render.Service{},
	},
	"v1/serviceaccounts": {
		Renderer: &render.ServiceAccount{},
	},
	"v1/persistentvolumes": {
		Renderer: &render.PersistentVolume{},
	},
	"v1/persistentvolumeclaims": {
		Renderer: &render.PersistentVolumeClaim{},
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
	"networking.k8s.io/v1/networkpolicies": {
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	"batch/v1beta1/cronjobs": {
		Renderer: &render.CronJob{},
	},
	"batch/v1/jobs": {
		DAO:      &dao.Job{},
		Renderer: &render.Job{},
	},

	// Autoscaling...
	"autoscaling/v1/horizontalpodautoscalers": {
		DAO:      &dao.HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},
	"autoscaling/v2beta1/horizontalpodautoscalers": {
		DAO:      &dao.HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},
	"autoscaling/v2beta2/horizontalpodautoscalers": {
		DAO:      &dao.HorizontalPodAutoscaler{},
		Renderer: &render.HorizontalPodAutoscaler{},
	},

	// CRDs...
	"apiextensions.k8s.io/v1/customresourcedefinitions": {
		DAO:      &dao.CustomResourceDefinition{},
		Renderer: &render.CustomResourceDefinition{},
	},
	"apiextensions.k8s.io/v1beta1/customresourcedefinitions": {
		DAO:      &dao.CustomResourceDefinition{},
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
		DAO:      &dao.Rbac{},
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
