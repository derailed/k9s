package model

import (
	"context"

	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// Context represents a kube context model.
type Context struct {
	Resource
}

// List returns a collection of node resources.
func (c *Context) List(_ context.Context) ([]runtime.Object, error) {
	cfg := c.factory.Client().Config()
	ctxs, err := cfg.Contexts()
	if err != nil {
		return nil, err
	}
	cc := make([]runtime.Object, 0, len(ctxs))
	for name, ctx := range ctxs {
		cc = append(cc, render.NewNamedContext(cfg, name, ctx))
	}

	return cc, nil
}
