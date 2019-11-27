package model

import (
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Resource represents a generic resource model.
type Resource struct {
	namespace, gvr string
	factory        *watch.Factory
}

func NewResource() *Resource {
	return &Resource{}
}

// NewResource returns a new model.
func (r *Resource) Init(ns, gvr string, f *watch.Factory) {
	r.namespace, r.gvr, r.factory = ns, gvr, f
}

// List returns a collection of nodes.
func (r *Resource) List(_ string) ([]runtime.Object, error) {
	return r.factory.List(r.namespace, r.gvr, labels.Everything())
}

// Render returns a node as a row.
func (r *Resource) Hydrate(oo []runtime.Object, rr render.Rows, re Renderer) error {
	var index int
	size := len(re.Header(r.namespace))
	for _, o := range oo {
		res := o.(*unstructured.Unstructured)
		row := render.Row{Fields: make([]string, size)}
		if err := re.Render(res, r.namespace, &row); err != nil {
			return err
		}
		rr[index] = row
		index++
	}

	return nil
}
