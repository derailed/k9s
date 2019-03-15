package resource

import (
	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
)

// SwitchableRes represents a resource that can be switched.
type SwitchableRes interface {
	k8s.ContextRes
}

// Context tracks a kubernetes resource.
type Context struct {
	*Base
	instance *k8s.NamedContext
}

// NewContextList returns a new resource list.
func NewContextList(ns string) List {
	return NewContextListWithArgs(ns, NewContext())
}

// NewContextListWithArgs returns a new resource list.
func NewContextListWithArgs(ns string, res Resource) List {
	return newList(NotNamespaced, "ctx", res, SwitchAccess)
}

// NewContext instantiates a new Context.
func NewContext() *Context {
	return NewContextWithArgs(k8s.NewContext().(SwitchableRes))
}

// NewContextWithArgs instantiates a new Context.
func NewContextWithArgs(r SwitchableRes) *Context {
	ctx := &Context{
		Base: &Base{
			caller: r,
		},
	}
	ctx.creator = ctx
	return ctx
}

// NewInstance builds a new Context instance from a k8s resource.
func (r *Context) NewInstance(i interface{}) Columnar {
	c := NewContext()
	switch i.(type) {
	case *k8s.NamedContext:
		c.instance = i.(*k8s.NamedContext)
	case k8s.NamedContext:
		ii := i.(k8s.NamedContext)
		c.instance = &ii
	default:
		log.Fatal().Msgf("unknown context type %#v", i)
	}
	c.path = c.instance.Name
	return c
}

// Switch out current context.
func (r *Context) Switch(c string) error {
	return r.caller.(k8s.ContextRes).Switch(c)
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

	name := i.Name
	if i.MustCurrentContextName() == name {
		name += "*"
	}

	return append(ff,
		name,
		i.Context.Cluster,
		i.Context.AuthInfo,
		i.Context.Namespace,
	)
}

// ExtFields returns extended fields in relation to headers.
func (*Context) ExtFields() Properties {
	return Properties{}
}
