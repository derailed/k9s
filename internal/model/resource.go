package model

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource represents a generic resource model.
type Resource struct {
	namespace, gvr string
	factory        dao.Factory
}

// Init initializes the model.
func (r *Resource) Init(ns, gvr string, f dao.Factory) {
	r.namespace, r.gvr, r.factory = ns, gvr, f
}

// List returns a collection of nodes.
func (r *Resource) List(ctx context.Context) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}

	return r.factory.List(r.gvr, r.namespace, true, lsel)
}

// Hydrate renders all rows.
func (r *Resource) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	for i, o := range oo {
		if err := re.Render(o, r.namespace, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}
