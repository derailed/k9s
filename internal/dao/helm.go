package dao

import (
	"context"
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"helm.sh/helm/v3/pkg/action"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor  = (*Helm)(nil)
	_ Nuker     = (*Helm)(nil)
	_ Describer = (*Helm)(nil)
)

// Helm represents a helm chart.
type Helm struct {
	NonResource
}

// List returns a collection of resources.
func (c *Helm) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}

	rr, err := action.NewList(cfg).Run()
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(rr))
	for _, r := range rr {
		oo = append(oo, render.HelmRes{Release: r})
	}

	return oo, nil
}

// Get returns a resource.
func (c *Helm) Get(_ context.Context, path string) (runtime.Object, error) {
	ns, n := client.Namespaced(path)
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return nil, err
	}

	return render.HelmRes{Release: resp}, nil
}

// Describe returns the chart notes.
func (c *Helm) Describe(path string) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return "", err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return "", err
	}

	return resp.Info.Notes, nil
}

// ToYAML returns the chart manifest.
func (c *Helm) ToYAML(path string, showManaged bool) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return "", err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return "", err
	}

	return resp.Manifest, nil
}

// Delete uninstall a Helm.
func (c *Helm) Delete(path string, cascade, force bool) error {
	ns, n := client.Namespaced(path)
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return err
	}

	res, err := action.NewUninstall(cfg).Run(n)
	if err != nil {
		return err
	}

	if res != nil && res.Info != "" {
		return fmt.Errorf("%s", res.Info)
	}

	return nil
}

// EnsureHelmConfig return a new configuration.
func (c *Helm) EnsureHelmConfig(ns string) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	if err := cfg.Init(c.Client().Config().Flags(), ns, os.Getenv("HELM_DRIVER"), helmLogger); err != nil {
		return nil, err
	}
	return cfg, nil
}

func helmLogger(s string, args ...interface{}) {
	log.Debug().Msgf("%s %v", s, args)
}
