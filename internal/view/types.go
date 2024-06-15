// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
)

const (
	ageCol      = "AGE"
	nameCol     = "NAME"
	statusCol   = "STATUS"
	cpuCol      = "CPU"
	memCol      = "MEM"
	uptodateCol = "UP-TO-DATE"
	readyCol    = "READY"
	availCol    = "AVAILABLE"
)

type (
	// EnvFunc represent the current view exposed environment.
	EnvFunc func() Env

	// BoostActionsFunc extends viewer keyboard actions.
	BoostActionsFunc func(ui.KeyActions)

	// EnterFunc represents an enter key action.
	EnterFunc func(app *App, model ui.Tabular, gvr client.GVR, path string)

	// LogOptionsFunc returns the active log options.
	LogOptionsFunc func(bool) (*dao.LogOptions, error)

	// ContextFunc enhances a given context.
	ContextFunc func(context.Context) context.Context

	// BindKeysFunc adds new menu actions.
	BindKeysFunc func(*ui.KeyActions)
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
	Actions() *ui.KeyActions

	// App returns an app handle.
	App() *App

	// Refresh updates the viewer
	Refresh()
}

// TableViewer represents a tabular viewer.
type TableViewer interface {
	Viewer

	// GetTable returns a table component.
	GetTable() *Table
}

// ResourceViewer represents a generic resource viewer.
type ResourceViewer interface {
	TableViewer

	// SetEnvFn sets a function to pull viewer env vars for plugins.
	SetEnvFn(EnvFunc)

	// GVR returns a resource descriptor.
	GVR() client.GVR

	// SetContextFn provision a custom context.
	SetContextFn(ContextFunc)

	// AddBindKeysFn provision additional key bindings.
	AddBindKeysFn(BindKeysFunc)

	// SetInstance sets a parent FQN
	SetInstance(string)
}

// LogViewer represents a log viewer.
type LogViewer interface {
	ResourceViewer

	ShowLogs(prev bool)
}

// RestartableViewer represents a viewer with restartable resources.
type RestartableViewer interface {
	LogViewer
}

// ScalableViewer represents a viewer with scalable resources.
type ScalableViewer interface {
	LogViewer
}

// SubjectViewer represents a policy viewer.
type SubjectViewer interface {
	ResourceViewer

	// SetSubject sets the active subject.
	SetSubject(s string)
}

// ViewerFunc returns a viewer matching a given gvr.
type ViewerFunc func(client.GVR) ResourceViewer

// MetaViewer represents a registered meta viewer.
type MetaViewer struct {
	viewerFn ViewerFunc
	enterFn  EnterFunc
}

// MetaViewers represents a collection of meta viewers.
type MetaViewers map[client.GVR]MetaViewer
