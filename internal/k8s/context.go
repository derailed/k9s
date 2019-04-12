package k8s

import (
	"fmt"

	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

// NamedContext represents a named cluster context.
type NamedContext struct {
	Name    string
	Context *api.Context
	config  *Config
}

// NewNamedContext returns a new named context.
func NewNamedContext(c *Config, n string, ctx *api.Context) *NamedContext {
	return &NamedContext{Name: n, Context: ctx, config: c}
}

// MustCurrentContextName return the active context name.
func (c *NamedContext) MustCurrentContextName() string {
	cl, err := c.config.CurrentContextName()
	if err != nil {
		panic(err)
	}
	return cl
}

// ----------------------------------------------------------------------------

// Context represents a Kubernetes Context.
type Context struct {
	*base
	Connection
}

// NewContext returns a new Context.
func NewContext(c Connection) *Context {
	return &Context{&base{}, c}
}

// Get a Context.
func (c *Context) Get(_, n string) (interface{}, error) {
	ctx, err := c.Config().GetContext(n)
	if err != nil {
		return nil, err
	}
	return &NamedContext{Name: n, Context: ctx}, nil
}

// List all Contexts on the current cluster.
func (c *Context) List(string) (Collection, error) {
	ctxs, err := c.Config().Contexts()
	if err != nil {
		return nil, err
	}
	cc := make([]interface{}, 0, len(ctxs))
	for k, v := range ctxs {
		cc = append(cc, NewNamedContext(c.Config(), k, v))
	}

	return cc, nil
}

// Delete a Context.
func (c *Context) Delete(_, n string) error {
	ctx, err := c.Config().CurrentContextName()
	if err != nil {
		return err
	}
	if ctx == n {
		return fmt.Errorf("trying to delete your current context %s", n)
	}
	return c.Config().DelContext(n)
}

// MustCurrentContextName return the active context name.
func (c *Context) MustCurrentContextName() string {
	cl, err := c.Config().CurrentContextName()
	if err != nil {
		panic(err)
	}
	return cl
}

// Switch to another context.
func (c *Context) Switch(ctx string) error {
	c.SwitchContextOrDie(ctx)
	return nil
}

// KubeUpdate modifies kubeconfig default context.
func (c *Context) KubeUpdate(n string) error {
	config, err := c.Config().RawConfig()
	if err != nil {
		return err
	}
	c.Switch(n)
	return clientcmd.ModifyConfig(
		clientcmd.NewDefaultPathOptions(), config, true,
	)
}
