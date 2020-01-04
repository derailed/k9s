package dao

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

var (
	_ Accessor = (*PortForward)(nil)
	_ Nuker    = (*PortForward)(nil)
)

// PortForward represents a port forward dao.
type PortForward struct {
	NonResource
}

// Delete a portforward.
func (p *PortForward) Delete(path string, cascade, force bool) error {
	ns, _ := client.Namespaced(path)
	auth, err := p.Client().CanI(ns, "v1/pods:portforward", []string{"delete"})
	if !auth || err != nil {
		return err
	}
	p.Factory.DeleteForwarder(path)

	return nil
}

// List returns a collection of screen dumps.
func (p *PortForward) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	config, ok := ctx.Value(internal.KeyBenchCfg).(*config.Bench)
	if !ok {
		return nil, fmt.Errorf("no benchconfig found in context")
	}

	cc := config.Benchmarks.Containers
	oo := make([]runtime.Object, 0, len(p.Factory.Forwarders()))
	for _, f := range p.Factory.Forwarders() {
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

// ----------------------------------------------------------------------------
// Helpers...

// ContainerID computes container ID based on ns/po/co.
func containerID(path, co string) string {
	ns, n := client.Namespaced(path)
	po := strings.Split(n, "-")[0]

	return ns + "/" + po + ":" + co
}
