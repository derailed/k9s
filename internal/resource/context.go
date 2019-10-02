package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
)

type (
	// Switchable represents a switchable resource.
	Switchable interface {
		Switch(ctx string) error
		MustCurrentContextName() string
	}

	// SwitchableCruder represents a resource that can be switched.
	SwitchableCruder interface {
		Cruder
		Switchable
	}

	// Context tracks a kubernetes resource.
	Context struct {
		*Base
		instance *k8s.NamedContext
	}
)

// NewContextList returns a new resource list.
func NewContextList(c Connection, _ string, _ k8s.GVR) List {
	return NewList(NotNamespaced, "ctx", NewContext(c), SwitchAccess)
}

// NewContext instantiates a new Context.
func NewContext(c Connection) *Context {
	ctx := &Context{Base: NewBase(c, k8s.NewContext(c))}
	ctx.Factory = ctx

	return ctx
}

// New builds a new Context instance from a k8s resource.
func (r *Context) New(i interface{}) Columnar {
	c := NewContext(r.Connection)
	switch instance := i.(type) {
	case *k8s.NamedContext:
		c.instance = instance
	case k8s.NamedContext:
		c.instance = &instance
	default:
		log.Fatal().Msgf("unknown context type %#v", i)
	}
	c.path = c.instance.Name

	return c
}

// Switch out current context.
func (r *Context) Switch(c string) error {
	return r.Resource.(Switchable).Switch(c)
}

// Marshal the resource to yaml.
func (r *Context) Marshal(path string) (string, error) {
	return "", nil
}

// Header return resource header.
func (*Context) Header(string) Row {
	return append(Row{}, "NAME", "CLUSTER", "AUTHINFO", "NAMESPACE")
}

// Fields retrieves displayable fields.
func (r *Context) Fields(ns string) Row {
	ff := make(Row, 0, len(r.Header(ns)))

	i := r.instance
	if i.MustCurrentContextName() == i.Name {
		i.Name += "*"
	}

	return append(ff,
		i.Name,
		i.Context.Cluster,
		i.Context.AuthInfo,
		i.Context.Namespace,
	)
}
