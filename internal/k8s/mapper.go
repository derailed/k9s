package k8s

import (
	"fmt"
	"os/user"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery/cached/disk"
	"k8s.io/client-go/restmapper"
)

var (
	// RestMapping holds k8s resource mapping
	// BOZO!! Has to be a better way...
	RestMapping = &RestMapper{}
	toFileName  = regexp.MustCompile(`[^(\w/\.)]`)
)

// RestMapper map resource to REST mapping ie kind, group, version.
type RestMapper struct {
	Connection
}

// Find a mapping given a resource kind.
func (*RestMapper) Find(kind string) (*meta.RESTMapping, error) {
	if m, ok := kindToMapper[kind]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("no mapping for kind %s", kind)
}

// ToRESTMapper map resources to kind, and map kind and version to interfaces for manipulating K8s objects.
func (r *RestMapper) ToRESTMapper() (meta.RESTMapper, error) {
	rc := r.RestConfigOrDie()

	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(rc.Host))

	disc, err := disk.NewCachedDiscoveryClientForConfig(rc, discCacheDir, httpCacheDir, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(disc)
	expander := restmapper.NewShortcutExpander(mapper, disc)
	return expander, nil
}

func toHostDir(host string) string {
	h := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	return toFileName.ReplaceAllString(h, "_")
}

func mustHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		panic(err)
	}
	return usr.HomeDir
}

// ResourceFor produces a rest mapping from a given resource.
// Support full res name ie deployment.v1.apps.
func (r *RestMapper) ResourceFor(resourceArg string) (*meta.RESTMapping, error) {
	res, err := r.resourceFor(resourceArg)
	if err != nil {
		return nil, err
	}
	return r.toRESTMapping(res, resourceArg), nil
}

func (r *RestMapper) resourceFor(resourceArg string) (schema.GroupVersionResource, error) {
	if resourceArg == "*" {
		return schema.GroupVersionResource{Resource: resourceArg}, nil
	}

	var (
		gvr schema.GroupVersionResource
		err error
	)

	mapper, err := r.ToRESTMapper()
	if err != nil {
		return gvr, err
	}

	fullGVR, gr := schema.ParseResourceArg(strings.ToLower(resourceArg))
	if fullGVR != nil {
		return mapper.ResourceFor(*fullGVR)
	}

	gvr, err = mapper.ResourceFor(gr.WithVersion(""))
	if err != nil {
		if len(gr.Group) == 0 {
			return gvr, fmt.Errorf("the server doesn't have a resource type '%s'", gr.Resource)
		}
		return gvr, fmt.Errorf("the server doesn't have a resource type '%s' in group '%s'", gr.Resource, gr.Group)
	}
	return gvr, nil
}

func (*RestMapper) toRESTMapping(gvr schema.GroupVersionResource, res string) *meta.RESTMapping {
	return &meta.RESTMapping{
		Resource:         gvr,
		GroupVersionKind: schema.GroupVersionKind{Group: gvr.Group, Version: gvr.Version, Kind: res},
		Scope:            RestMapping,
	}
}

// Name protocol returns rest scope name.
func (*RestMapper) Name() meta.RESTScopeName {
	return meta.RESTScopeNameNamespace
}

var kindToMapper = map[string]*meta.RESTMapping{
	"ConfigMap": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmap"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		Scope:            RestMapping,
	},
	"Pod": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pod"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		Scope:            RestMapping,
	},
	"Service": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "service"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		Scope:            RestMapping,
	},
	"EndPoints": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Endpoints"},
		Scope:            RestMapping,
	},
	"Namespace": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespace"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		Scope:            RestMapping,
	},
	"Node": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "node"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
		Scope:            RestMapping,
	},
	"PersistentVolume": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolume"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolume"},
		Scope:            RestMapping,
	},
	"PersistentVolumeClaim": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaim"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		Scope:            RestMapping,
	},
	"ReplicationController": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "replicationcontroller"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"},
		Scope:            RestMapping,
	},
	"Secret": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secret"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		Scope:            RestMapping,
	},
	"StorageClasse": {
		Resource:         schema.GroupVersionResource{Group: "storage.k8s.io", Version: "v1", Resource: "storageclass"},
		GroupVersionKind: schema.GroupVersionKind{Group: "storage.k8s.io", Version: "v1", Kind: "StorageClass"},
		Scope:            RestMapping,
	},
	"ServiceAccount": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccount"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		Scope:            RestMapping,
	},

	"Deployment": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployment"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Scope:            RestMapping,
	},
	"ReplicaSet": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicaset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
		Scope:            RestMapping,
	},
	"StatefulSet": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		Scope:            RestMapping,
	},

	"HorizontalPodAutoscaler": {
		Resource:         schema.GroupVersionResource{Group: "autoscaling", Version: "v1", Resource: "horizontalpodautoscaler"},
		GroupVersionKind: schema.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "HorizontalPodAutoscaler"},
		Scope:            RestMapping,
	},

	"Job": {
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "job"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		Scope:            RestMapping,
	},
	"CronJob": {
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjob"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"},
		Scope:            RestMapping,
	},

	"DaemonSet": {
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "daemonset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet"},
		Scope:            RestMapping,
	},
	"Ingress": {
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingress"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Ingress"},
		Scope:            RestMapping,
	},

	"ClusterRole": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrole"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Scope:            RestMapping,
	},
	"ClusterRoleBinding": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		Scope:            RestMapping,
	},
	"Role": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "role"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		Scope:            RestMapping,
	},
	"RoleBinding": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		Scope:            RestMapping,
	},

	"CustomResourceDefinition": {
		Resource:         schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"},
		Scope:            RestMapping,
	},
	"NetworkPolicy": {
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "networkpolicies"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "NetworkPolicy"},
		Scope:            RestMapping,
	},

	"Event": {
		Resource:         schema.GroupVersionResource{Group: "events.k8s.io", Version: "v1beta1", Resource: "events"},
		GroupVersionKind: schema.GroupVersionKind{Group: "events.k8s.io", Version: "v1beta1", Kind: "Event"},
		Scope:            RestMapping,
	},

	"PodDisruptionBudget": {
		Resource:         schema.GroupVersionResource{Group: "policy", Version: "v1beta1", Resource: "poddisruptionbudgets"},
		GroupVersionKind: schema.GroupVersionKind{Group: "policy", Version: "v1beta1", Kind: "PodDisruptionBudget"},
		Scope:            RestMapping,
	},
}
