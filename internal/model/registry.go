package model

import (
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/xray"
)

// Registry tracks resources metadata.
// BOZO!! Break up deps and merge into single registrar.
var Registry = map[string]ResourceMeta{
	// Custom...
	"references": {
		DAO:      &dao.Reference{},
		Renderer: &render.Reference{},
	},
	"dir": {
		DAO:      &dao.Dir{},
		Renderer: &render.Dir{},
	},
	"pulses": {
		DAO: &dao.Pulse{},
	},
	"helm": {
		DAO:      &dao.Helm{},
		Renderer: &render.Helm{},
	},
	// BOZO!! revamp with latest...
	// "openfaas": {
	// 	DAO:      &dao.OpenFaas{},
	// 	Renderer: &render.OpenFaas{},
	// },
	"containers": {
		DAO:          &dao.Container{},
		Renderer:     &render.Container{},
		TreeRenderer: &xray.Container{},
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
	"popeye": {
		DAO:      &dao.Popeye{},
		Renderer: &render.Popeye{},
	},
	"sanitizer": {
		DAO:          &dao.Popeye{},
		TreeRenderer: &xray.Section{},
	},

	// Core...
	"v1/endpoints": {
		Renderer: &render.Endpoints{},
	},
	"v1/pods": {
		DAO:          &dao.Pod{},
		Renderer:     &render.Pod{},
		TreeRenderer: &xray.Pod{},
	},
	"v1/namespaces": {
		Renderer: &render.Namespace{},
	},
	"v1/nodes": {
		DAO:      &dao.Node{},
		Renderer: &render.Node{},
	},
	"v1/services": {
		DAO:          &dao.Service{},
		Renderer:     &render.Service{},
		TreeRenderer: &xray.Service{},
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
	"v1/events": {
		DAO:      &dao.Table{},
		Renderer: &render.Event{},
	},

	// Apps...
	"apps/v1/deployments": {
		DAO:          &dao.Deployment{},
		Renderer:     &render.Deployment{},
		TreeRenderer: &xray.Deployment{},
	},
	"apps/v1/replicasets": {
		Renderer:     &render.ReplicaSet{},
		TreeRenderer: &xray.ReplicaSet{},
	},
	"apps/v1/statefulsets": {
		DAO:          &dao.StatefulSet{},
		Renderer:     &render.StatefulSet{},
		TreeRenderer: &xray.StatefulSet{},
	},
	"apps/v1/daemonsets": {
		DAO:          &dao.DaemonSet{},
		Renderer:     &render.DaemonSet{},
		TreeRenderer: &xray.DaemonSet{},
	},

	// Extensions...
	"networking.k8s.io/v1/networkpolicies": {
		Renderer: &render.NetworkPolicy{},
	},

	// Batch...
	"batch/v1/cronjobs": {
		DAO:      &dao.CronJob{},
		Renderer: &render.CronJob{},
	},
	"batch/v1/jobs": {
		DAO:      &dao.Job{},
		Renderer: &render.Job{},
	},

	// CRDs...
	"apiextensions.k8s.io/v1/customresourcedefinitions": {
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
