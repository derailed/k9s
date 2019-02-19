package k8s

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd/api"
)

// ContextRes represents a kubernetes clusters configurations.
type ContextRes interface {
	Res
	Switch(n string) error
}

// NamedContext represents a named cluster context.
type NamedContext struct {
	Name    string
	Context *api.Context
}

// MustCurrentClusterName return the active cluster name.
func (c *NamedContext) MustCurrentClusterName() string {
	cl, err := conn.config.CurrentClusterName()
	if err != nil {
		panic(err)
	}
	return cl
}

// Context represents a Kubernetes Context.
type Context struct{}

// NewContext returns a new Context.
func NewContext() Res {
	return &Context{}
}

// Get a Context.
func (*Context) Get(_, n string) (interface{}, error) {
	ctx, err := conn.config.GetContext(n)
	if err != nil {
		return nil, err
	}
	return &NamedContext{Name: n, Context: ctx}, nil
}

// List all Contexts in a given namespace
func (*Context) List(string) (Collection, error) {
	ctxs, err := conn.config.Contexts()
	if err != nil {
		return Collection{}, err
	}
	cc := make([]interface{}, 0, len(ctxs))
	for k, v := range ctxs {
		cc = append(cc, &NamedContext{k, v})
	}
	return cc, nil
}

// Delete a Context
func (*Context) Delete(_, n string) error {
	ctx, err := conn.config.CurrentContextName()
	if err != nil {
		return err
	}
	if ctx == n {
		return fmt.Errorf("trying to delete your current context %s", n)
	}
	return conn.config.DelContext(n)
}

// Switch cluster Context.
func (*Context) Switch(n string) error {
	conn.switchContextOrDie(n)
	return nil
}
