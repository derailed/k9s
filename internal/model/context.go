package model

import (
	"context"
	"fmt"

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

// Hydrate returns nodes as rows.
func (n *Context) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	var index int
	for _, o := range oo {
		ctx, ok := o.(*render.NamedContext)
		if !ok {
			return fmt.Errorf("expecting named context but got %T", o)
		}

		var row render.Row
		if err := re.Render(ctx, "", &row); err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}
