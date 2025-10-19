// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"fmt"
	"log/slog"
	"path"
	"strings"
	"sync"

	"github.com/derailed/k9s/internal/slogs"
	"github.com/fvbommel/sortorder"
	"gopkg.in/yaml.v3"
	apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var NoGVR = new(GVR)

// GVR represents a kubernetes resource schema as a string.
// Format is group/version/resources:subresource.
type GVR struct {
	raw, g, v, r, sr string
}

type gvrCache struct {
	data map[string]*GVR
	sync.RWMutex
}

func (c *gvrCache) add(gvr *GVR) {
	if c.get(gvr.String()) == nil {
		c.Lock()
		c.data[gvr.String()] = gvr
		c.Unlock()
	}
}

func (c *gvrCache) get(gvrs string) *GVR {
	c.RLock()
	defer c.RUnlock()

	if gvr, ok := c.data[gvrs]; ok {
		return gvr
	}

	return nil
}

var gvrsCache = gvrCache{
	data: make(map[string]*GVR),
}

// NewGVR builds a new gvr from a group, version, resource.
func NewGVR(s string) *GVR {
	raw := s
	tokens := strings.Split(s, ":")
	var g, v, r, sr string
	if len(tokens) == 2 {
		raw, sr = tokens[0], tokens[1]
	}
	tokens = strings.Split(raw, "/")
	switch len(tokens) {
	case 3:
		g, v, r = tokens[0], tokens[1], tokens[2]
	case 2:
		v, r = tokens[0], tokens[1]
	case 1:
		r = tokens[0]
	default:
		slog.Error("GVR init failed!", slogs.Error, fmt.Errorf("can't parse GVR %q", s))
	}

	gvr := GVR{raw: s, g: g, v: v, r: r, sr: sr}
	if cgvr := gvrsCache.get(gvr.String()); cgvr != nil {
		return cgvr
	}
	gvrsCache.add(&gvr)

	return &gvr
}

func (g *GVR) IsAlias() bool {
	return !g.IsK8sRes()
}

func (g *GVR) IsK8sRes() bool {
	return g != nil && ((!strings.Contains(g.raw, " ") && strings.Contains(g.raw, "/") && !strings.Contains(g.raw, " /")) || reservedGVRs.Has(g))
}

// WithSubResource builds a new gvr with a sub resource.
func (g *GVR) WithSubResource(sub string) *GVR {
	return NewGVR(g.String() + ":" + sub)
}

// NewGVRFromMeta builds a gvr from resource metadata.
func NewGVRFromMeta(a *metav1.APIResource) *GVR {
	return NewGVR(path.Join(a.Group, a.Version, a.Name))
}

// NewGVRFromCRD builds a gvr from a custom resource definition.
func NewGVRFromCRD(crd *apiext.CustomResourceDefinition) map[*GVR]*apiext.CustomResourceDefinitionVersion {
	mm := make(map[*GVR]*apiext.CustomResourceDefinitionVersion, len(crd.Spec.Versions))
	for _, v := range crd.Spec.Versions {
		if v.Served && !v.Deprecated {
			gvr := NewGVRFromMeta(&metav1.APIResource{
				Kind:    crd.Spec.Names.Kind,
				Group:   crd.Spec.Group,
				Name:    crd.Spec.Names.Plural,
				Version: v.Name,
			})
			mm[gvr] = &v
		}
	}

	return mm
}

// FromGVAndR builds a gvr from a group/version and resource.
func FromGVAndR(gv, r string) *GVR {
	return NewGVR(path.Join(gv, r))
}

// FQN returns a fully qualified resource name.
func (g *GVR) FQN(n string) string {
	return path.Join(g.AsResourceName(), n)
}

// AsResourceName returns a resource . separated descriptor in the shape of kind.version.group.
func (g *GVR) AsResourceName() string {
	if g.g == "" {
		return g.r
	}

	return g.r + "." + g.v + "." + g.g
}

// SubResource returns a sub resource if available.
func (g *GVR) SubResource() string {
	return g.sr
}

// String returns gvr as string.
func (g *GVR) String() string {
	return g.raw
}

// GV returns the group version scheme representation.
func (g *GVR) GV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.g,
		Version: g.v,
	}
}

// GVK returns a full schema representation.
func (g *GVR) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.G(),
		Version: g.V(),
		Kind:    g.R(),
	}
}

// GVR returns a full schema representation.
func (g *GVR) GVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.G(),
		Version:  g.V(),
		Resource: g.R(),
	}
}

// GVSub returns group vervion sub path.
func (g *GVR) GVSub() string {
	if g.G() == "" {
		return g.V()
	}

	return g.G() + "/" + g.V()
}

// GR returns a full schema representation.
func (g *GVR) GR() *schema.GroupResource {
	return &schema.GroupResource{
		Group:    g.G(),
		Resource: g.R(),
	}
}

// V returns the resource version.
func (g *GVR) V() string {
	return g.v
}

// RG returns the resource and group.
func (g *GVR) RG() (resource, group string) {
	return g.r, g.g
}

// R returns the resource name.
func (g *GVR) R() string {
	return g.r
}

// G returns the resource group name.
func (g *GVR) G() string {
	return g.g
}

// IsDecodable checks if the k8s resource has a decodable view
func (g *GVR) IsDecodable() bool {
	return g == SecGVR
}

var _ = yaml.Marshaler((*GVR)(nil))
var _ = yaml.Unmarshaler((*GVR)(nil))

func (g *GVR) MarshalYAML() (any, error) {
	return g.String(), nil
}

func (g *GVR) UnmarshalYAML(n *yaml.Node) error {
	*g = *NewGVR(n.Value)

	return nil
}

// GVRs represents a collection of gvr.
type GVRs []*GVR

// Len returns the list size.
func (g GVRs) Len() int {
	return len(g)
}

// Swap swaps list values.
func (g GVRs) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

// Less returns true if i < j.
func (g GVRs) Less(i, j int) bool {
	g1, g2 := g[i].G(), g[j].G()

	return sortorder.NaturalLess(g1, g2)
}

// Helper...

// Can determines the available actions for a given resource.
func Can(verbs []string, v string) bool {
	if verbs == nil {
		return true
	}
	if len(verbs) == 0 {
		return false
	}
	for _, verb := range verbs {
		candidates, err := mapVerb(v)
		if err != nil {
			slog.Error("Access verb mapping failed", slogs.Error, err)
			return false
		}
		for _, c := range candidates {
			if verb == c {
				return true
			}
		}
	}

	return false
}

func mapVerb(v string) ([]string, error) {
	switch v {
	case "describe":
		return []string{"get"}, nil
	case "view":
		return []string{"get", "list"}, nil
	case "delete":
		return []string{"delete"}, nil
	case "edit":
		return []string{"patch", "update"}, nil
	default:
		return []string{}, fmt.Errorf("no standard verb for %q", v)
	}
}
