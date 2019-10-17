package views

import (
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/watch"
)

func forwardColorer(string, *resource.RowEvent) tcell.Color {
	return tcell.ColorSkyblue
}

func dumpColorer(ns string, r *resource.RowEvent) tcell.Color {
	return tcell.ColorNavajoWhite
}

func benchColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := tcell.ColorPaleGreen

	statusCol := 2
	if strings.TrimSpace(r.Fields[statusCol]) != "pass" {
		c = ui.ErrColor
	}

	return c
}

func aliasColorer(string, *resource.RowEvent) tcell.Color {
	return tcell.ColorMediumSpringGreen
}

func rbacColorer(ns string, r *resource.RowEvent) tcell.Color {
	return ui.DefaultColorer(ns, r)
}

func podColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	readyCol := 2
	if len(ns) != 0 {
		readyCol = 1
	}
	statusCol := readyCol + 1

	tokens := strings.Split(strings.TrimSpace(r.Fields[readyCol]), "/")
	if len(tokens) == 2 && (tokens[0] == "0" || tokens[0] != tokens[1]) {
		if strings.TrimSpace(r.Fields[statusCol]) != "Completed" {
			c = ui.ErrColor
		}
	}

	switch strings.TrimSpace(r.Fields[statusCol]) {
	case "ContainerCreating", "PodInitializing":
		return ui.AddColor
	case "Initialized":
		return ui.HighlightColor
	case "Completed":
		return ui.CompletedColor
	case "Running":
	default:
		c = ui.ErrColor
	}

	return c
}

func containerColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	readyCol := 2
	if strings.TrimSpace(r.Fields[readyCol]) == "false" {
		c = ui.ErrColor
	}

	stateCol := readyCol + 1
	switch strings.TrimSpace(r.Fields[stateCol]) {
	case "ContainerCreating", "PodInitializing":
		return ui.AddColor
	case "Terminating", "Initialized":
		return ui.HighlightColor
	case "Completed":
		return ui.CompletedColor
	case "Running":
	default:
		c = ui.ErrColor
	}

	return c
}

func ctxColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	if strings.Contains(strings.TrimSpace(r.Fields[0]), "*") {
		c = ui.HighlightColor
	}

	return c
}

func pvColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	status := strings.TrimSpace(r.Fields[4])
	switch status {
	case "Bound":
		c = ui.StdColor
	case "Available":
		c = tcell.ColorYellow
	default:
		c = ui.ErrColor
	}

	return c
}

func pvcColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}

	if strings.TrimSpace(r.Fields[markCol]) != "Bound" {
		c = ui.ErrColor
	}

	return c
}

func pdbColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 5
	if ns != resource.AllNamespaces {
		markCol = 4
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func dpColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func stsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func rsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func evColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	markCol := 3
	if ns != resource.AllNamespaces {
		markCol = 2
	}

	switch strings.TrimSpace(r.Fields[markCol]) {
	case "Failed":
		c = ui.ErrColor
	case "Killing":
		c = ui.KillColor
	}

	return c
}

func nsColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	switch strings.TrimSpace(r.Fields[1]) {
	case "Inactive", "Terminating":
		c = ui.ErrColor
	}

	if strings.Contains(strings.TrimSpace(r.Fields[0]), "*") {
		c = ui.HighlightColor
	}

	return c
}
