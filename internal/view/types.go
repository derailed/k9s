package view

import (
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type (
	// EnvFunc represent the current view exposed environment.
	EnvFunc func() K9sEnv

	// BoostActionFunc extends viewer keyboard actions.
	BoostActionsFunc func(ui.KeyActions)

	// ViewFunc represents a new resource viewer.
	ViewFunc func(title, gvr string, list resource.List) ResourceViewer

	// ListFunc represents a new resource list.
	ListFunc func(c resource.Connection, ns string) resource.List

	// EnterFunc represents an enter key action.
	EnterFunc func(app *App, ns, resource, selection string)

	// DecorateFunc represents a row decorator.
	DecorateFunc func(resource.TableData) resource.TableData

	// ContainerFunc returns the active container name.
	ContainerFunc func() string
)

// ActionExtender enhances a given viewer by adding new menu actions.
type ActionExtender interface {
	// BindKeys injects new menu actions.
	BindKeys(ResourceViewer)
}

// Hinter represents a view that can produce menu hints.
type Hinter interface {
	// Hints returns a collection of hints.
	Hints() model.MenuHints
}

// Viewer represents a component viewer.
type Viewer interface {
	model.Component

	// Actions returns active menu bindings.
	Actions() ui.KeyActions

	// App returns an app handle.
	App() *App

	// Refresh updates the viewer
	Refresh()
}

// ResourceViewer represents a generic resource viewer.
type ResourceViewer interface {
	TableViewer

	// List returns a resource List.
	List() resource.List

	// SetEnvFn sets a function to pull viewer env vars for plugins.
	SetEnvFn(EnvFunc)
}

// TableViewer represents a tabular viewer.
type TableViewer interface {
	Viewer

	// Table returns a table component.
	GetTable() *Table
}

type LogViewer interface {
	ResourceViewer

	ShowLogs(prev bool)
}

type RestartableViewer interface {
	LogViewer
}

type ScalableViewer interface {
	LogViewer
}

// SubjectViewer represents a policy viewer.
type SubjectViewer interface {
	ResourceViewer

	// SetSubject sets the active subject.
	SetSubject(s string)
}

// MetaViewer represents a registered meta viewer.
type MetaViewer struct {
	gvr        string
	kind       string
	namespaced bool
	verbs      metav1.Verbs
	viewFn     ViewFunc
	listFn     ListFunc
	enterFn    EnterFunc
	colorerFn  ui.ColorerFunc
	decorateFn DecorateFunc
}

// MetaViewers represents a collection of meta viewers.
type MetaViewers map[string]MetaViewer
