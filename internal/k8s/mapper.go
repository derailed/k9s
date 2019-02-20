package k8s

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// RestMapping holds k8s resource mapping
// BOZO!! Has to be a better way...
var RestMapping = &RestMapper{}

// RestMapper map resource to REST mapping ie kind, group, version.
type RestMapper struct{}

// Find a mapping given a resource name.
func (*RestMapper) Find(res string) (*meta.RESTMapping, error) {
	if m, ok := resMap[res]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("no mapping for resource %s", res)
}

// Name protocol returns rest scope name.
func (*RestMapper) Name() meta.RESTScopeName {
	return meta.RESTScopeNameNamespace
}

var resMap = map[string]*meta.RESTMapping{
	"ConfigMaps": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmap"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		Scope:            RestMapping,
	},
	"Pods": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pod"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		Scope:            RestMapping,
	},
	"Services": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "service"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		Scope:            RestMapping,
	},
	"EndPoints": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Endpoints"},
		Scope:            RestMapping,
	},
	"Namespaces": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespace"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		Scope:            RestMapping,
	},
	"Nodes": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "node"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
		Scope:            RestMapping,
	},
	"PersistentVolumes": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolume"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolume"},
		Scope:            RestMapping,
	},
	"PersistentVolumeClaims": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaim"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		Scope:            RestMapping,
	},
	"ReplicationControllers": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "replicationcontroller"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"},
		Scope:            RestMapping,
	},
	"Secrets": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secret"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		Scope:            RestMapping,
	},
	"ServiceAccounts": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccount"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		Scope:            RestMapping,
	},

	"Deployments": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployment"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Scope:            RestMapping,
	},
	"ReplicaSets": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicaset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
		Scope:            RestMapping,
	},
	"StatefulSets": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		Scope:            RestMapping,
	},

	"HorizontalPodAutoscalers": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "autoscaling", Version: "v1", Resource: "horizontalpodautoscaler"},
		GroupVersionKind: schema.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "HorizontalPodAutoscaler"},
		Scope:            RestMapping,
	},

	"Jobs": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "job"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		Scope:            RestMapping,
	},
	"CronJobs": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjob"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"},
		Scope:            RestMapping,
	},

	"DaemonSets": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "daemonset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet"},
		Scope:            RestMapping,
	},
	"Ingress": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingress"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Ingress"},
		Scope:            RestMapping,
	},

	"ClusterRoles": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrole"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Scope:            RestMapping,
	},
	"ClusterRoleBindings": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		Scope:            RestMapping,
	},
	"Roles": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "role"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		Scope:            RestMapping,
	},
	"RoleBindings": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		Scope:            RestMapping,
	},

	"CustomResourceDefinitions": &meta.RESTMapping{
		Resource:         schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"},
		Scope:            RestMapping,
	},
}

// {Group: "certificates.k8s.io", Version: "v1beta1", Kind: "CertificateSigningRequest"}:     {},
// {Group: "certificates.k8s.io", Version: "v1beta1", Kind: "CertificateSigningRequestList"}: {},
// {Group: "kubeadm.k8s.io", Version: "v1alpha1", Kind: "MasterConfiguration"}:               {},
// {Group: "extensions", Version: "v1beta1", Kind: "PodSecurityPolicy"}:                                    {},
// {Group: "extensions", Version: "v1beta1", Kind: "PodSecurityPolicyList"}:                                {},
// {Group: "extensions", Version: "v1beta1", Kind: "NetworkPolicy"}:                                        {},
// {Group: "extensions", Version: "v1beta1", Kind: "NetworkPolicyList"}:                                    {},
// {Group: "policy", Version: "v1beta1", Kind: "PodSecurityPolicy"}:                                        {},
// {Group: "policy", Version: "v1beta1", Kind: "PodSecurityPolicyList"}:                                    {},
// {Group: "settings.k8s.io", Version: "v1alpha1", Kind: "PodPreset"}:                                      {},
// {Group: "settings.k8s.io", Version: "v1alpha1", Kind: "PodPresetList"}:                                  {},
// {Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "ValidatingWebhookConfiguration"}:     {},
// {Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "ValidatingWebhookConfigurationList"}: {},
// {Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "MutatingWebhookConfiguration"}:       {},
// {Group: "admissionregistration.k8s.io", Version: "v1beta1", Kind: "MutatingWebhookConfigurationList"}:   {},
// {Group: "auditregistration.k8s.io", Version: "v1alpha1", Kind: "AuditSink"}:                             {},
// {Group: "auditregistration.k8s.io", Version: "v1alpha1", Kind: "AuditSinkList"}:                         {},
// {Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicy"}:                                      {},
// {Group: "networking.k8s.io", Version: "v1", Kind: "NetworkPolicyList"}:                                  {},
// {Group: "storage.k8s.io", Version: "v1beta1", Kind: "StorageClass"}:                                     {},
// {Group: "storage.k8s.io", Version: "v1beta1", Kind: "StorageClassList"}:                                 {},
// {Group: "storage.k8s.io", Version: "v1", Kind: "StorageClass"}:                                          {},
// {Group: "storage.k8s.io", Version: "v1", Kind: "StorageClassList"}:                                      {},
// {Group: "authentication.k8s.io", Version: "v1", Kind: "TokenRequest"}:                                   {},
