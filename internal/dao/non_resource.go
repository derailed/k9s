package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// NonResource represents a non k8s resource.
type NonResource struct {
	Factory

	gvr client.GVR
}

// Init initializes the resource.
func (g *NonResource) Init(f Factory, gvr client.GVR) {
	g.Factory, g.gvr = f, gvr
}

func (g *NonResource) GVR() string {
	return g.gvr.String()
}

// Get returns the given resource.
func (c *NonResource) Get(context.Context, string) (runtime.Object, error) {
	return nil, fmt.Errorf("NYI!")
}
