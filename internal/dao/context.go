package dao

import (
	"fmt"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/rs/zerolog/log"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Context struct {
	Generic
}

var _ Accessor = &Context{}
var _ Switchable = &Context{}

func (c *Context) config() *k8s.Config {
	return c.Factory.Client().Config()
}

// Get a Context.
func (c *Context) Get(_, n string) (runtime.Object, error) {
	ctx, err := c.config().GetContext(n)
	if err != nil {
		return nil, err
	}
	return &NamedContext{Name: n, Context: ctx}, nil
}

// List all Contexts on the current cluster.
func (c *Context) List(string, metav1.ListOptions) ([]runtime.Object, error) {
	ctxs, err := c.config().Contexts()
	if err != nil {
		return nil, err
	}
	cc := make([]runtime.Object, 0, len(ctxs))
	for k, v := range ctxs {
		cc = append(cc, NewNamedContext(c.config(), k, v))
	}

	return cc, nil
}

// Delete a Context.
func (c *Context) Delete(path string, cascade, force bool) error {
	ctx, err := c.config().CurrentContextName()
	if err != nil {
		return err
	}
	if ctx == path {
		return fmt.Errorf("trying to delete your current context %s", path)
	}

	return c.config().DelContext(path)
}

// MustCurrentContextName return the active context name.
func (c *Context) MustCurrentContextName() string {
	cl, err := c.config().CurrentContextName()
	if err != nil {
		log.Fatal().Err(err).Msg("Fetching current context")
	}
	return cl
}

// Switch to another context.
func (c *Context) Switch(ctx string) error {
	c.Factory.Client().SwitchContextOrDie(ctx)
	return nil
}

// KubeUpdate modifies kubeconfig default context.
func (c *Context) KubeUpdate(n string) error {
	config, err := c.config().RawConfig()
	if err != nil {
		return err
	}
	if err := c.Switch(n); err != nil {
		return err
	}
	return clientcmd.ModifyConfig(
		clientcmd.NewDefaultPathOptions(), config, true,
	)
}

// ----------------------------------------------------------------------------

// NamedContext represents a named cluster context.
type NamedContext struct {
	Name    string
	Context *api.Context
	config  *k8s.Config
}

// NewNamedContext returns a new named context.
func NewNamedContext(c *k8s.Config, n string, ctx *api.Context) *NamedContext {
	return &NamedContext{Name: n, Context: ctx, config: c}
}

// MustCurrentContextName return the active context name.
func (c *NamedContext) MustCurrentContextName() string {
	cl, err := c.config.CurrentContextName()
	if err != nil {
		log.Fatal().Err(err).Msg("Fetching current context")
	}
	return cl
}

// GetObjectKind returns a schema object.
func (c *NamedContext) GetObjectKind() schema.ObjectKind {
	return nil
}

// DeepCopyObject returns a container copy.
func (c *NamedContext) DeepCopyObject() runtime.Object {
	return c
}
