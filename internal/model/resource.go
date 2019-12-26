package model

import (
	"context"
	"time"

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
	defer func(t time.Time) {
		log.Debug().Msgf("LIST elapsed: %v", time.Since(t))
	}(time.Now())

	strLabel, ok := ctx.Value(internal.KeyLabels).(string)
	lsel := labels.Everything()
	if sel, err := labels.ConvertSelectorToLabelsMap(strLabel); ok && err == nil {
		lsel = sel.AsSelector()
	}
	return r.factory.List(r.gvr, r.namespace, lsel)
}

// Render returns a node as a row.
func (r *Resource) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	defer func(t time.Time) {
		log.Debug().Msgf("HYDRATE elapsed: %v", time.Since(t))
	}(time.Now())

	for i, o := range oo {
		if err := re.Render(o, r.namespace, &rr[i]); err != nil {
			return err
		}
	}

	return nil
}
