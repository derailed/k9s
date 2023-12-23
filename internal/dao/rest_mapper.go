// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/restmapper"
)

// RestMapping holds k8s resource mapping.
var RestMapping = &RestMapper{}

// RestMapper map resource to REST mapping ie kind, group, version.
type RestMapper struct {
	client.Connection
}

// ToRESTMapper map resources to kind, and map kind and version to interfaces for manipulating K8s objects.
func (r *RestMapper) ToRESTMapper() (meta.RESTMapper, error) {
	dial, err := r.CachedDiscovery()
	if err != nil {
		return nil, err
	}
	mapper := restmapper.NewDeferredDiscoveryRESTMapper(dial)
	expander := restmapper.NewShortcutExpander(mapper, dial, nil)

	return expander, nil
}

// ResourceFor produces a rest mapping from a given resource.
// Support full res name ie deployment.v1.apps.
func (r *RestMapper) ResourceFor(resourceArg, kind string) (*meta.RESTMapping, error) {
	res, err := r.resourceFor(resourceArg)
	if err != nil {
		return nil, err
	}
	return r.toRESTMapping(res, kind), nil
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

func (*RestMapper) toRESTMapping(gvr schema.GroupVersionResource, kind string) *meta.RESTMapping {
	return &meta.RESTMapping{
		Resource: gvr,
		GroupVersionKind: schema.GroupVersionKind{
			Group:   gvr.Group,
			Version: gvr.Version,
			Kind:    kind,
		},
		Scope: RestMapping,
	}
}

// Name protocol returns rest scope name.
func (*RestMapper) Name() meta.RESTScopeName {
	return meta.RESTScopeNameNamespace
}
