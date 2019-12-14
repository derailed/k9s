package model

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/render"
	"github.com/rs/zerolog/log"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource represents a generic resource model.
type Resource struct {
	namespace, gvr string
	factory        Factory
}

func (r *Resource) Init(ns, gvr string, f Factory) {
	r.namespace, r.gvr, r.factory = ns, gvr, f
}

// List returns a collection of nodes.
func (r *Resource) List(ctx context.Context) ([]runtime.Object, error) {
	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}
	log.Debug().Msgf("^^^^^Listing with selector %q:%q--%#v", r.namespace, r.gvr, lsel)
	oo, err := r.factory.List(r.gvr, r.namespace, lsel)
	r.factory.WaitForCacheSync()

	return oo, err
}

// Render returns a node as a row.
func (r *Resource) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	log.Debug().Msgf("^^^^^^ HYDRATING (%q) %d", r.namespace, len(oo))

	var index int
	for _, o := range oo {
		var row render.Row
		if err := re.Render(o, r.namespace, &row); err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}
