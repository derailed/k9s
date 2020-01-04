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
	_ Accessor  = (*Chart)(nil)
	_ Nuker     = (*Chart)(nil)
	_ Describer = (*Chart)(nil)
)

// Chart represents a helm chart.
type Chart struct {
	NonResource
}

// List returns a collection of resources.
func (c *Chart) List(ctx context.Context, ns string) ([]runtime.Object, error) {
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
		oo = append(oo, render.ChartRes{Release: r})
	}

	return oo, nil
}

// Get returns a resource.
func (c *Chart) Get(_ context.Context, path string) (runtime.Object, error) {
	ns, n := client.Namespaced(path)
	cfg, err := c.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return nil, err
	}

	return render.ChartRes{Release: resp}, nil
}

// Describe returns the chart notes.
func (c *Chart) Describe(path string) (string, error) {
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
func (c *Chart) ToYAML(path string) (string, error) {
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

// Delete uninstall a Chart.
func (c *Chart) Delete(path string, cascade, force bool) error {
	log.Debug().Msgf("CHART DELETE %q", path)
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

func (c *Chart) EnsureHelmConfig(ns string) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	flags := c.Client().Config().Flags()
	if err := cfg.Init(flags, ns, os.Getenv("HELM_DRIVER"), helmLogger); err != nil {
		return nil, err
	}
	return cfg, nil
}

func helmLogger(s string, args ...interface{}) {
	log.Debug().Msgf("%s %v", s, args)
}
