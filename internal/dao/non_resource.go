package dao

import (
	"context"
	"fmt"
	"sync"

	"github.com/derailed/k9s/internal/client"
	"k8s.io/apimachinery/pkg/runtime"
)

// NonResource represents a non k8s resource.
type NonResource struct {
	Factory

	gvr client.GVR
	mx  sync.RWMutex
}

// Init initializes the resource.
func (n *NonResource) Init(f Factory, gvr client.GVR) {
	n.mx.Lock()
	defer n.mx.Unlock()
	n.Factory, n.gvr = f, gvr
}

// GVR returns a gvr.
func (n *NonResource) GVR() string {
	n.mx.RLock()
	defer n.mx.RUnlock()
	return n.gvr.String()
}

// Get returns the given resource.
func (n *NonResource) Get(context.Context, string) (runtime.Object, error) {
	return nil, fmt.Errorf("NYI!")
}
