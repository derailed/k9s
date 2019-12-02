package k8s

import (
	"path"
	"strings"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

// GVR represents a fully qualified kubernetes resource.
type GVR string

// NewGVR returns a new gvr.
func NewGVR(g, v, r string) GVR {
	return GVR(path.Join(g, v, r))
}

// ToGVR returns a new gvr from a string.
func ToGVR(gv, r string) GVR {
	return GVR(path.Join(gv, r))
}

// ResName returns a full res name ie dp.v1.apps.
func (g GVR) ResName() string {
	return g.ToR() + "." + g.ToV() + "." + g.ToG()
}

// AsGV returns the group version.
func (g GVR) AsGV() schema.GroupVersion {
	return schema.GroupVersion{
		Group:   g.ToG(),
		Version: g.ToV(),
	}
}

// AsGVR returns a schema gvr instance.
func (g GVR) AsGVR() schema.GroupVersionResource {
	return schema.GroupVersionResource{
		Group:    g.ToG(),
		Version:  g.ToV(),
		Resource: g.ToR(),
	}
}

// String returns a GVR as a string.
func (g GVR) String() string {
	return string(g)
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
