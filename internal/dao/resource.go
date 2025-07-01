// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

var (
	_ Accessor  = (*Resource)(nil)
	_ Describer = (*Resource)(nil)
	_ Nuker     = (*Resource)(nil)
)

// Resource represents an informer based resource.
type Resource struct {
	Generic
}

// List returns a collection of resources.
func (r *Resource) List(ctx context.Context, ns string) ([]runtime.Object, error) {
	lsel := labels.Everything()
	if sel, ok := ctx.Value(internal.KeyLabels).(labels.Selector); ok {
		lsel = sel
	}

	return r.getFactory().List(r.gvr, ns, false, lsel)
}

// Get returns a resource instance if found, else an error.
func (r *Resource) Get(_ context.Context, path string) (runtime.Object, error) {
	return r.getFactory().Get(r.gvr, path, true, labels.Everything())
}

// ToYAML returns a resource yaml.
func (r *Resource) ToYAML(path string, showManaged bool) (string, error) {
	o, err := r.Get(context.Background(), path)
	if err != nil {
		return "", err
	}

	raw, err := ToYAML(o, showManaged)
	if err != nil {
		return "", fmt.Errorf("unable to marshal resource %w", err)
	}
	return raw, nil
}
