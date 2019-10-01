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
