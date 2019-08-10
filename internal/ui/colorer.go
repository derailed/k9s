package ui

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	// ModColor row modified color.
	ModColor tcell.Color
	// AddColor row added color.
	AddColor tcell.Color
	// ErrColor row err color.
	ErrColor tcell.Color
	// StdColor row default color.
	StdColor tcell.Color
	// HighlightColor row highlight color.
	HighlightColor tcell.Color
	// KillColor row deleted color.
	KillColor tcell.Color
	// CompletedColor row completed color.
	CompletedColor tcell.Color
)

// DefaultColorer set the default table row colors.
func DefaultColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := StdColor
	switch r.Action {
	case watch.Added, resource.New:
		c = AddColor
	case watch.Modified:
		c = ModColor
	}
	return c
}
