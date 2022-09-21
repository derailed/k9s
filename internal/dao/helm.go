package dao

import (
	"context"
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
func (h *Helm) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	cfg, err := h.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}

	list := action.NewList(cfg)
	list.All = true
	list.SetStateMask()
	rr, err := list.Run()
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
func (h *Helm) Get(_ context.Context, path string) (runtime.Object, error) {
	ns, n := client.Namespaced(path)
	cfg, err := h.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return nil, err
	}

	return render.HelmRes{Release: resp}, nil
}

// GetValues returns values for a release
func (h *Helm) GetValues(path string, allValues bool) ([]byte, error) {
	ns, n := client.Namespaced(path)
	cfg, err := h.EnsureHelmConfig(ns)
	if err != nil {
		return nil, err
	}
	vals := action.NewGetValues(cfg)
	vals.AllValues = allValues
	resp, err := vals.Run(n)
	if err != nil {
		return nil, err
	}

	return yaml.Marshal(resp)
}

// Describe returns the chart notes.
func (h *Helm) Describe(path string) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := h.EnsureHelmConfig(ns)
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
func (h *Helm) ToYAML(path string, showManaged bool) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := h.EnsureHelmConfig(ns)
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
func (h *Helm) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, force bool) error {
	ns, n := client.Namespaced(path)
	cfg, err := h.EnsureHelmConfig(ns)
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
func (h *Helm) EnsureHelmConfig(ns string) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	err := cfg.Init(h.Client().Config().Flags(), ns, os.Getenv("HELM_DRIVER"), helmLogger)

	return cfg, err
}

func helmLogger(s string, args ...interface{}) {
	log.Debug().Msgf("%s %v", s, args)
}
