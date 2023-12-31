// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"os"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render/helm"
	"github.com/rs/zerolog/log"
	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor  = (*HelmChart)(nil)
	_ Nuker     = (*HelmChart)(nil)
	_ Describer = (*HelmChart)(nil)
	_ Valuer    = (*HelmChart)(nil)
)

// HelmChart represents a helm chart.
type HelmChart struct {
	NonResource
}

// List returns a collection of resources.
func (h *HelmChart) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	cfg, err := ensureHelmConfig(h.Client(), ns)
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
		oo = append(oo, helm.ReleaseRes{Release: r})
	}

	return oo, nil
}

// Get returns a resource.
func (h *HelmChart) Get(_ context.Context, path string) (runtime.Object, error) {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
	if err != nil {
		return nil, err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return nil, err
	}

	return helm.ReleaseRes{Release: resp}, nil
}

// GetValues returns values for a release
func (h *HelmChart) GetValues(path string, allValues bool) ([]byte, error) {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
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
func (h *HelmChart) Describe(path string) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
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
func (h *HelmChart) ToYAML(path string, showManaged bool) (string, error) {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
	if err != nil {
		return "", err
	}
	resp, err := action.NewGet(cfg).Run(n)
	if err != nil {
		return "", err
	}

	return resp.Manifest, nil
}

// Delete uninstall a HelmChart.
func (h *HelmChart) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, _ Grace) error {
	return h.Uninstall(path, false)
}

// Uninstall uninstalls a HelmChart.
func (h *HelmChart) Uninstall(path string, keepHist bool) error {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
	if err != nil {
		return err
	}
	u := action.NewUninstall(cfg)
	u.KeepHistory = keepHist
	res, err := u.Run(n)
	if err != nil {
		return err
	}
	if res != nil && res.Info != "" {
		return fmt.Errorf("%s", res.Info)
	}

	return nil
}

// ensureHelmConfig return a new configuration.
func ensureHelmConfig(c client.Connection, ns string) (*action.Configuration, error) {
	cfg := new(action.Configuration)
	err := cfg.Init(c.Config().Flags(), ns, os.Getenv("HELM_DRIVER"), helmLogger)

	return cfg, err
}

func helmLogger(s string, args ...interface{}) {
	log.Debug().Msgf("%s %v", s, args)
}
