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
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/restmapper"
)

// RestMapping holds k8s resource mapping
// BOZO!! Has to be a better way...
var RestMapping = &RestMapper{}

// RestMapper map resource to REST mapping ie kind, group, version.
type RestMapper struct{}

// Find a mapping given a resource name.
func (*RestMapper) Find1(res string) (*meta.RESTMapping, error) {
	if m, ok := resMap[res]; ok {
		return m, nil
	}
	return nil, fmt.Errorf("no mapping for resource %s", res)
}

func (*RestMapper) ToRESTMapper() (meta.RESTMapper, error) {
	rc := conn.restConfigOrDie()

	httpCacheDir := filepath.Join(mustHomeDir(), ".kube", "http-cache")
	discCacheDir := filepath.Join(mustHomeDir(), ".kube", "cache", "discovery", toHostDir(rc.Host))

	disc, err := discovery.NewCachedDiscoveryClientForConfig(rc, discCacheDir, httpCacheDir, 10*time.Minute)
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(disc)
	expander := restmapper.NewShortcutExpander(mapper, disc)
	return expander, nil
}

var toFileName = regexp.MustCompile(`[^(\w/\.)]`)

func toHostDir(host string) string {
	h := strings.Replace(strings.Replace(host, "https://", "", 1), "http://", "", 1)
	// now do a simple collapse of non-AZ09 characters.  Collisions are possible but unlikely.  Even if we do collide the problem is short lived
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

var resMap = map[string]*meta.RESTMapping{
	"ConfigMaps": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "configmap"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ConfigMap"},
		Scope:            RestMapping,
	},
	"Pods": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "pod"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Pod"},
		Scope:            RestMapping,
	},
	"Services": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "service"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Service"},
		Scope:            RestMapping,
	},
	"EndPoints": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "endpoints"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Endpoints"},
		Scope:            RestMapping,
	},
	"Namespaces": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespace"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Namespace"},
		Scope:            RestMapping,
	},
	"Nodes": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "node"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Node"},
		Scope:            RestMapping,
	},
	"PersistentVolumes": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolume"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolume"},
		Scope:            RestMapping,
	},
	"PersistentVolumeClaims": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "persistentvolumeclaim"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "PersistentVolumeClaim"},
		Scope:            RestMapping,
	},
	"ReplicationControllers": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "replicationcontroller"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ReplicationController"},
		Scope:            RestMapping,
	},
	"Secrets": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "secret"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "Secret"},
		Scope:            RestMapping,
	},
	"ServiceAccounts": {
		Resource:         schema.GroupVersionResource{Group: "", Version: "v1", Resource: "serviceaccount"},
		GroupVersionKind: schema.GroupVersionKind{Group: "", Version: "v1", Kind: "ServiceAccount"},
		Scope:            RestMapping,
	},

	"Deployments": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployment"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "Deployment"},
		Scope:            RestMapping,
	},
	"ReplicaSets": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "replicaset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "ReplicaSet"},
		Scope:            RestMapping,
	},
	"StatefulSets": {
		Resource:         schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "statefulsets"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apps", Version: "v1", Kind: "StatefulSet"},
		Scope:            RestMapping,
	},

	"HorizontalPodAutoscalers": {
		Resource:         schema.GroupVersionResource{Group: "autoscaling", Version: "v1", Resource: "horizontalpodautoscaler"},
		GroupVersionKind: schema.GroupVersionKind{Group: "autoscaling", Version: "v1", Kind: "HorizontalPodAutoscaler"},
		Scope:            RestMapping,
	},

	"Jobs": {
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "job"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "Job"},
		Scope:            RestMapping,
	},
	"CronJobs": {
		Resource:         schema.GroupVersionResource{Group: "batch", Version: "v1", Resource: "cronjob"},
		GroupVersionKind: schema.GroupVersionKind{Group: "batch", Version: "v1", Kind: "CronJob"},
		Scope:            RestMapping,
	},

	"DaemonSets": {
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "daemonset"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "DaemonSet"},
		Scope:            RestMapping,
	},
	"Ingress": {
		Resource:         schema.GroupVersionResource{Group: "extensions", Version: "v1beta1", Resource: "ingress"},
		GroupVersionKind: schema.GroupVersionKind{Group: "extensions", Version: "v1beta1", Kind: "Ingress"},
		Scope:            RestMapping,
	},

	"ClusterRoles": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrole"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRole"},
		Scope:            RestMapping,
	},
	"ClusterRoleBindings": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "clusterrolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "ClusterRoleBinding"},
		Scope:            RestMapping,
	},
	"Roles": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "role"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "Role"},
		Scope:            RestMapping,
	},
	"RoleBindings": {
		Resource:         schema.GroupVersionResource{Group: "rbac.authorization.k8s.io", Version: "v1", Resource: "rolebinding"},
		GroupVersionKind: schema.GroupVersionKind{Group: "rbac.authorization.k8s.io", Version: "v1", Kind: "RoleBinding"},
		Scope:            RestMapping,
	},

	"CustomResourceDefinitions": {
		Resource:         schema.GroupVersionResource{Group: "apiextensions.k8s.io", Version: "v1beta1", Resource: "customresourcedefinitions"},
		GroupVersionKind: schema.GroupVersionKind{Group: "apiextensions.k8s.io", Version: "v1beta1", Kind: "CustomResourceDefinition"},
		Scope:            RestMapping,
	},
}
