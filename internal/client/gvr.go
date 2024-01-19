// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package client

import (
	"fmt"
	"path"
	"strings"

	"github.com/fvbommel/sortorder"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var NoGVR = GVR{}

// GVR represents a kubernetes resource schema as a string.
// Format is group/version/resources:subresource.
type GVR struct {
	raw, g, v, r, sr string
}

// NewGVR builds a new gvr from a group, version, resource.
func NewGVR(gvr string) GVR {
	var g, v, r, sr string

	tokens := strings.Split(gvr, ":")
	raw := gvr
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
		log.Error().Err(fmt.Errorf("can't parse GVR %q", gvr)).Msg("GVR init failed!")
	}

	return GVR{raw: gvr, g: g, v: v, r: r, sr: sr}
}

// NewGVRFromMeta builds a gvr from resource metadata.
func NewGVRFromMeta(a metav1.APIResource) GVR {
	return GVR{
		raw: path.Join(a.Group, a.Version, a.Name),
		g:   a.Group,
		v:   a.Version,
		r:   a.Name,
	}
}

// FromGVAndR builds a gvr from a group/version and resource.
func FromGVAndR(gv, r string) GVR {
	return NewGVR(path.Join(gv, r))
}

// FQN returns a fully qualified resource name.
func (g GVR) FQN(n string) string {
	return path.Join(g.AsResourceName(), n)
}

// AsResourceName returns a resource . separated descriptor in the shape of kind.version.group.
func (g GVR) AsResourceName() string {
	return g.r + "." + g.v + "." + g.g
}

// SubResource returns a sub resource if available.
func (g GVR) SubResource() string {
	return g.sr
}

// String returns gvr as string.
func (g GVR) String() string {
	return g.raw
}

// GV returns the group version scheme representation.
func (g GVR) GV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.g,
		Version: g.v,
	}
}

// GVK returns a full schema representation.
func (g GVR) GVK() schema.GroupVersionKind {
	return schema.GroupVersionKind{
		Group:   g.G(),
		Version: g.V(),
		Kind:    g.R(),
	}
}

// GVR returns a full schema representation.
func (g GVR) GVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.G(),
		Version:  g.V(),
		Resource: g.R(),
	}
}

// GR returns a full schema representation.
func (g GVR) GR() *schema.GroupResource {
	return &schema.GroupResource{
		Group:    g.G(),
		Resource: g.R(),
	}
}

// V returns the resource version.
func (g GVR) V() string {
	return g.v
}

// RG returns the resource and group.
func (g GVR) RG() (string, string) {
	return g.r, g.g
}

// R returns the resource name.
func (g GVR) R() string {
	return g.r
}

// G returns the resource group name.
func (g GVR) G() string {
	return g.g
}

// IsDecodable checks if the k8s resource has a decodable view
func (g GVR) IsDecodable() bool {
	return g.GVK().Kind == "secrets"
}

// GVRs represents a collection of gvr.
type GVRs []GVR

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
			log.Error().Err(err).Msgf("verb mapping failed")
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
