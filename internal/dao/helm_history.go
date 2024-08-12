// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
	"helm.sh/helm/v3/pkg/action"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render/helm"
)

var (
	_ Accessor  = (*HelmHistory)(nil)
	_ Nuker     = (*HelmHistory)(nil)
	_ Describer = (*HelmHistory)(nil)
	_ Valuer    = (*HelmHistory)(nil)
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

	cfg, err := ensureHelmConfig(h.Client().Config().Flags(), ns)
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
	fqn, rev, found := strings.Cut(path, ":")
	if !found || len(rev) == 0 {
		return nil, fmt.Errorf("invalid path %q", path)
	}

	ns, n := client.Namespaced(fqn)
	cfg, err := ensureHelmConfig(h.Client().Config().Flags(), ns)
	if err != nil {
		return nil, err
	}

	getter := action.NewGet(cfg)
	getter.Version, err = strconv.Atoi(rev)
	if err != nil {
		return nil, err
	}

	resp, err := getter.Run(n)
	if err != nil {
		return nil, err
	}

	return helm.ReleaseRes{Release: resp}, nil
}

// Describe returns the chart notes.
func (h *HelmHistory) Describe(path string) (string, error) {
	rel, err := h.Get(context.Background(), path)
	if err != nil {
		return "", err
	}

	resp, ok := rel.(helm.ReleaseRes)
	if !ok {
		return "", fmt.Errorf("expected helm.ReleaseRes, but got %T", rel)
	}

	return resp.Release.Info.Notes, nil
}

// ToYAML returns the chart manifest.
func (h *HelmHistory) ToYAML(path string, showManaged bool) (string, error) {
	rel, err := h.Get(context.Background(), path)
	if err != nil {
		return "", err
	}

	resp, ok := rel.(helm.ReleaseRes)
	if !ok {
		return "", fmt.Errorf("expected helm.ReleaseRes, but got %T", rel)
	}

	return resp.Release.Manifest, nil
}

// GetValues return the config for this chart.
func (h *HelmHistory) GetValues(path string, allValues bool) ([]byte, error) {
	rel, err := h.Get(context.Background(), path)
	if err != nil {
		return nil, err
	}

	resp, ok := rel.(helm.ReleaseRes)
	if !ok {
		return nil, fmt.Errorf("expected helm.ReleaseRes, but got %T", rel)
	}

	if allValues {
		return yaml.Marshal(resp.Release.Chart.Values)
	}
	return yaml.Marshal(resp.Release.Config)
}

func (h *HelmHistory) Rollback(_ context.Context, path, rev string) error {
	ns, n := client.Namespaced(path)
	cfg, err := ensureHelmConfig(h.Client().Config().Flags(), ns)
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
	cfg, err := ensureHelmConfig(h.Client().Config().Flags(), ns)
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
