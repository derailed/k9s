package views

import (
	"strings"

	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	modColor       tcell.Color
	addColor       tcell.Color
	errColor       tcell.Color
	stdColor       tcell.Color
	highlightColor tcell.Color
	killColor      tcell.Color
	completedColor tcell.Color
)

func defaultColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := stdColor
	switch r.Action {
	case watch.Added, resource.New:
		c = addColor
	case watch.Modified:
		c = modColor
	}
	return c
}

func forwardColorer(string, *resource.RowEvent) tcell.Color {
	return tcell.ColorSkyblue
}

func benchColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := tcell.ColorPaleGreen

	statusCol := 2
	if strings.TrimSpace(r.Fields[statusCol]) != "pass" {
		c = errColor
	}

	return c
}

func aliasColorer(string, *resource.RowEvent) tcell.Color {
	return tcell.ColorFuchsia
}

func rbacColorer(ns string, r *resource.RowEvent) tcell.Color {
	return defaultColorer(ns, r)
}

func podColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)

	readyCol := 2
	if len(ns) != 0 {
		readyCol = 1
	}
	statusCol := readyCol + 1

	tokens := strings.Split(strings.TrimSpace(r.Fields[readyCol]), "/")
	if len(tokens) == 2 && (tokens[0] == "0" || tokens[0] != tokens[1]) {
		if strings.TrimSpace(r.Fields[statusCol]) != "Completed" {
			c = errColor
		}
	}

	switch strings.TrimSpace(r.Fields[statusCol]) {
	case "ContainerCreating", "PodInitializing":
		return addColor
	case "Terminating", "Initialized":
		return highlightColor
	case "Completed":
		return completedColor
	case "Running":
	default:
		c = errColor
	}

	return c
}

func containerColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)

	readyCol := 2
	if strings.TrimSpace(r.Fields[readyCol]) == "false" {
		c = errColor
	}

	stateCol := readyCol + 1
	switch strings.TrimSpace(r.Fields[stateCol]) {
	case "ContainerCreating", "PodInitializing":
		return addColor
	case "Terminating", "Initialized":
		return highlightColor
	case "Completed":
		return completedColor
	case "Running":
	default:
		c = errColor
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

	status := strings.TrimSpace(r.Fields[4])
	switch status {
	case "Bound":
		c = stdColor
	case "Available":
		c = tcell.ColorYellow
	default:
		c = errColor
	}

	return c
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

func pdbColorer(ns string, r *resource.RowEvent) tcell.Color {
	c := defaultColorer(ns, r)
	if r.Action == watch.Added || r.Action == watch.Modified {
		return c
	}

	markCol := 5
	if ns != resource.AllNamespaces {
		markCol = 4
	}
	if strings.TrimSpace(r.Fields[markCol]) != strings.TrimSpace(r.Fields[markCol+1]) {
		return errColor
	}

	return stdColor
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
