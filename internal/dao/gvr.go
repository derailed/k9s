package dao

import (
	"fmt"
	"path"
	"strings"

	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"vbom.ml/util/sortorder"
)

// GVR represents a kubernetes resource schema as a string.
// Format is group/version/resources
type GVR string

// NewGVR builds a new gvr from a group, version, resource.
func NewGVR(g, v, r string) GVR {
	return GVR(path.Join(g, v, r))
}

// FromGVAndR builds a gvr from a group/version and resource.
func FromGVAndR(gv, r string) GVR {
	return GVR(path.Join(gv, r))
}

// ResName returns a resource . separated descriptor in the shape of kind.version.group.
func (g GVR) ResName() string {
	return g.ToR() + "." + g.ToV() + "." + g.ToG()
}

// AsGV returns the group version scheme representation.
func (g GVR) AsGV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.ToG(),
		Version: g.ToV(),
	}
}

// AsGVR returns a a full schema representation.
func (g GVR) AsGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.ToG(),
		Version:  g.ToV(),
		Resource: g.ToR(),
	}
}

// ToV returns the resource version.
func (g GVR) ToV() string {
	tokens := strings.Split(string(g), "/")
	if len(tokens) < 2 {
		return ""
	}
	return tokens[len(tokens)-2]
}

// ToR returns the resource name.
func (g GVR) ToR() string {
	tokens := strings.Split(string(g), "/")
	return tokens[len(tokens)-1]
}

// ToG returns the resource group name.
func (g GVR) ToG() string {
	tokens := strings.Split(string(g), "/")
	switch len(tokens) {
	case 3:
		return tokens[0]
	default:
		return ""
	}
}

type GVRs []GVR

func (g GVRs) Len() int {
	return len(g)
}

func (g GVRs) Swap(i, j int) {
	g[i], g[j] = g[j], g[i]
}

func (g GVRs) Less(i, j int) bool {
	g1, g2 := g[i].ToG(), g[j].ToG()

	return sortorder.NaturalLess(g1, g2)
}

// Helper...

// Can determines the available actions for a given resource.
func Can(verbs []string, v string) bool {
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
