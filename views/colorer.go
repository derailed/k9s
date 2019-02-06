package views

import (
	"strings"

	"github.com/derailed/k9s/resource"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/watch"
)

const (
	modColor       = tcell.ColorGreenYellow
	addColor       = tcell.ColorLightSkyBlue
	errColor       = tcell.ColorOrangeRed
	stdColor       = tcell.ColorWhite
	highlightColor = tcell.ColorAqua
	killColor      = tcell.ColorMediumPurple
)

func defaultColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := stdColor
	switch r.Action {
	case watch.Added:
		c = addColor
	case watch.Modified:
		c = modColor
	}
	return c
}

func podColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)

	statusCol := 3
	if len(ns) != 0 {
		statusCol = 2
	}
	switch strings.TrimSpace(r.Fields[statusCol]) {
	case "Running", "Initialized", "Completed", "Terminating":
	default:
		c = errColor
	}

	readyCol := 2
	if len(ns) != 0 {
		readyCol = 1
	}
	tokens := strings.Split(strings.TrimSpace(r.Fields[readyCol]), "/")
	if len(tokens) == 2 && (tokens[0] == "0" || tokens[0] != tokens[1]) {
		if strings.TrimSpace(r.Fields[statusCol]) != "Completed" {
			c = errColor
		}
	}
	return c
}

func ctxColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	if strings.Contains(strings.TrimSpace(r.Fields[0]), "*") {
		c = highlightColor
	}
	return c
}

func pvColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	if strings.TrimSpace(r.Fields[4]) != "Bound" {
		return errColor
	}
	return stdColor
}

func pvcColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}

	if strings.TrimSpace(r.Fields[markCol]) != "Bound" {
		c = errColor
	}
	return c
}

func dpColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return errColor
	}
	return stdColor
}

func stsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return errColor
	}
	return stdColor
}

func rsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return errColor
	}
	return stdColor
}

func evColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)

	markCol := 3
	if ns != resource.AllNamespaces {
		markCol = 2
	}

	switch strings.TrimSpace(r.Fields[markCol]) {
	case "Failed":
		c = errColor
	case "Killing":
		c = killColor
	}
	return c
}

func nsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	switch strings.TrimSpace(r.Fields[1]) {
	case "Inactive", "Terminating":
		c = errColor
	}

	if strings.Contains(strings.TrimSpace(r.Fields[0]), "*") {
		c = highlightColor
	}

	return c
}
