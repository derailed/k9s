package model

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/runtime"
)

// PortForward represents a portforward model.
type PortForward struct {
	Resource
}

// List returns a collection of screen dumps.
func (c *PortForward) List(ctx context.Context) ([]runtime.Object, error) {
	config, ok := ctx.Value(internal.KeyBenchCfg).(*config.Bench)
	if !ok {
		return nil, fmt.Errorf("no benchconfig found in context")
	}

	cc := config.Benchmarks.Containers
	oo := make([]runtime.Object, 0, len(c.factory.Forwarders()))
	for _, f := range c.factory.Forwarders() {
		cfg := render.BenchCfg{
			C: config.Benchmarks.Defaults.C,
			N: config.Benchmarks.Defaults.N,
		}
		if config, ok := cc[containerID(f.Path(), f.Container())]; ok {
			cfg.C, cfg.N = config.C, config.N
			cfg.Host, cfg.Path = config.HTTP.Host, config.HTTP.Path
		}
		oo = append(oo, render.ForwardRes{
			Forwarder: f,
			Config:    cfg,
		})
	}

	return oo, nil
}

// Hydrate returns a pod as container rows.
func (c *PortForward) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		res, ok := o.(render.ForwardRes)
		if !ok {
			return fmt.Errorf("expecting a forwardres but got %T", o)
		}

		if err := re.Render(res, render.NonResource, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}

// ----------------------------------------------------------------------------
// Helpers...

// ContainerID computes container ID based on ns/po/co.
func containerID(path, co string) string {
	ns, n := client.Namespaced(path)
	po := strings.Split(n, "-")[0]

	return ns + "/" + po + ":" + co
}
