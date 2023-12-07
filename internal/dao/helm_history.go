// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"strconv"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render/helm"
	"helm.sh/helm/v3/pkg/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor  = (*HelmHistory)(nil)
	_ Nuker     = (*HelmHistory)(nil)
	_ Describer = (*HelmHistory)(nil)
)

// HelmHistory represents a helm chart.
type HelmHistory struct {
	NonResource
}

// List returns a collection of resources.
func (h *HelmHistory) List(ctx context.Context, _ string) ([]runtime.Object, error) {
	path, ok := ctx.Value(internal.KeyFQN).(string)
	if !ok {
		return nil, fmt.Errorf("expecting FQN in context")
	}
	ns, n := client.Namespaced(path)

	cfg, err := ensureHelmConfig(h.Client(), ns)
	if err != nil {
		return nil, err
	}

	hh, err := action.NewHistory(cfg).Run(n)
	if err != nil {
		return nil, err
	}

	oo := make([]runtime.Object, 0, len(hh))
	for _, r := range hh {
		oo = append(oo, helm.ReleaseRes{Release: r})
	}

	return oo, nil
}

// Get returns a resource.
func (h *HelmHistory) Get(_ context.Context, path string) (runtime.Object, error) {
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

// Describe returns the chart notes.
func (h *HelmHistory) Describe(path string) (string, error) {
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
func (h *HelmHistory) ToYAML(path string, showManaged bool) (string, error) {
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

func (h *HelmHistory) Rollback(_ context.Context, path, rev string) error {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
	if err != nil {
		return err
	}

	ver, err := strconv.Atoi(rev)
	if err != nil {
		return fmt.Errorf("could not convert revision to a number: %w", err)
	}
	client := action.NewRollback(cfg)
	client.Version = ver

	return client.Run(n)
}

// Delete uninstall a Helm.
func (h *HelmHistory) Delete(_ context.Context, path string, _ *metav1.DeletionPropagation, _ Grace) error {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client(), ns)
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
