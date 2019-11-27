package model

import (
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/watch"
	"k8s.io/apimachinery/pkg/runtime"
)

type Renderer interface {
	// Render converts raw resources to tabular data.
	Render(o interface{}, ns string, row *render.Row) error

	// Header returns the resource header.
	Header(ns string) render.HeaderRow

	ColorerFunc() render.ColorerFunc
}

type Lister interface {
	// Init initializes a resource.
	Init(ns, gvr string, f *watch.Factory)

	// List returns a collection of resources.
	List(sel string) ([]runtime.Object, error)

	// Hydrate converts resource rows into tabular data.
	Hydrate([]runtime.Object, render.Rows, Renderer) error
}

type ResourceMeta struct {
	Model    Lister
	Renderer Renderer
}

var Registry = map[string]ResourceMeta{
	"v1/pods": ResourceMeta{
		Model:    NewPod(),
		Renderer: &render.Pod{},
	},
	"v1/nodes": ResourceMeta{
		Model:    NewNode(),
		Renderer: &render.Node{},
	},
	"v1/configmaps": ResourceMeta{
		Model:    NewResource(),
		Renderer: &render.ConfigMap{},
	},
	"containers": ResourceMeta{
		Model:    NewContainer(),
		Renderer: &render.Container{},
	},
}
