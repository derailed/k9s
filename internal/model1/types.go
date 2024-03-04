// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package model1

import (
	"github.com/derailed/tcell/v2"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	NAValue = "na"

	// EventUnchanged notifies listener resource has not changed.
	EventUnchanged ResEvent = 1 << iota

	// EventAdd notifies listener of a resource was added.
	EventAdd

	// EventUpdate notifies listener of a resource updated.
	EventUpdate

	// EventDelete  notifies listener of a resource was deleted.
	EventDelete

	// EventClear the stack was reset.
	EventClear
)

// DecoratorFunc decorates a string.
type DecoratorFunc func(string) string

// ColorerFunc represents a resource row colorer.
type ColorerFunc func(ns string, h Header, re *RowEvent) tcell.Color

// Renderer represents a resource renderer.
type Renderer interface {
	// IsGeneric identifies a generic handler.
	IsGeneric() bool

	// Render converts raw resources to tabular data.
	Render(o interface{}, ns string, row *Row) error

	// Header returns the resource header.
	Header(ns string) Header

	// ColorerFunc returns a row colorer function.
	ColorerFunc() ColorerFunc
}

// Generic represents a generic resource.
type Generic interface {
	// SetTable sets up the resource tabular definition.
	SetTable(ns string, table *metav1.Table)

	// Header returns a resource header.
	Header(ns string) Header

	// Render renders the resource.
	Render(o interface{}, ns string, row *Row) error
}
