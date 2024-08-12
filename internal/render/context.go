// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd/api"
)

// Context renders a K8s ConfigMap to screen.
type Context struct {
	Base
}

// ColorerFunc colors a resource row.
func (Context) ColorerFunc() model1.ColorerFunc {
	return func(ns string, h model1.Header, r *model1.RowEvent) tcell.Color {
		c := model1.DefaultColorer(ns, h, r)
		if strings.Contains(strings.TrimSpace(r.Row.Fields[0]), "*") {
			return model1.HighlightColor
		}

		return c
	}
}

// Header returns a header row.
func (Context) Header(ns string) model1.Header {
	return model1.Header{
		model1.HeaderColumn{Name: "NAME"},
		model1.HeaderColumn{Name: "CLUSTER"},
		model1.HeaderColumn{Name: "AUTHINFO"},
		model1.HeaderColumn{Name: "NAMESPACE"},
	}
}

// Render renders a K8s resource to screen.
func (c Context) Render(o interface{}, _ string, r *model1.Row) error {
	ctx, ok := o.(*NamedContext)
	if !ok {
		return fmt.Errorf("expected *NamedContext, but got %T", o)
	}

	name := ctx.Name
	if ctx.IsCurrentContext(ctx.Name) {
		name += "(*)"
	}

	r.ID = ctx.Name
	r.Fields = model1.Fields{
		name,
		ctx.Context.Cluster,
		ctx.Context.AuthInfo,
		ctx.Context.Namespace,
	}

	return nil
}

// Helpers...

// NamedContext represents a named cluster context.
type NamedContext struct {
	Name    string
	Context *api.Context
	Config  ContextNamer
}

// ContextNamer represents a named context.
type ContextNamer interface {
	CurrentContextName() (string, error)
}

// NewNamedContext returns a new named context.
func NewNamedContext(c ContextNamer, n string, ctx *api.Context) *NamedContext {
	return &NamedContext{Name: n, Context: ctx, Config: c}
}

// IsCurrentContext return the active context name.
func (c *NamedContext) IsCurrentContext(n string) bool {
	cl, err := c.Config.CurrentContextName()
	if err != nil {
		log.Fatal().Err(err).Msg("Fetching current context")
		return false
	}
	return cl == n
}

// GetObjectKind returns a schema object.
func (c *NamedContext) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c *NamedContext) DeepCopyObject() runtime.Object {
	return c
}
