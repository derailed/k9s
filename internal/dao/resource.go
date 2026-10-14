// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dao

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/fields"
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

	oo, err := r.getFactory().List(r.gvr, ns, false, lsel)
	if err != nil {
		return nil, err
	}

	return filterByFieldSelector(ctx, oo)
}

// filterByFieldSelector applies a client-side field selector if present in the context.
func filterByFieldSelector(ctx context.Context, oo []runtime.Object) ([]runtime.Object, error) {
	sel, ok := ctx.Value(internal.KeyFields).(string)
	if !ok || sel == "" {
		return oo, nil
	}
	fsel, err := fields.ParseSelector(sel)
	if err != nil {
		return nil, err
	}

	res := make([]runtime.Object, 0, len(oo))
	for _, o := range oo {
		u, ok := o.(*unstructured.Unstructured)
		if !ok {
			// Cannot evaluate a field selector on a non-unstructured object.
			res = append(res, o)
			continue
		}
		set := make(fields.Set, len(fsel.Requirements()))
		for _, req := range fsel.Requirements() {
			val, _, _ := unstructured.NestedString(u.Object, strings.Split(req.Field, ".")...)
			set[req.Field] = val
		}
		if fsel.Matches(set) {
			res = append(res, o)
		}
	}

	return res, nil
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
