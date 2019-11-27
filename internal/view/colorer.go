package view

import (
	"strings"

	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

func forwardColorer(string, render.RowEvent) tcell.Color {
	return tcell.ColorSkyblue
}

func dumpColorer(ns string, r render.RowEvent) tcell.Color {
	return tcell.ColorNavajoWhite
}

func benchColorer(ns string, r render.RowEvent) tcell.Color {
	c := tcell.ColorPaleGreen

	statusCol := 2
	if strings.TrimSpace(r.Row.Fields[statusCol]) != "pass" {
		c = ui.ErrColor
	}

	return c
}

func aliasColorer(string, render.RowEvent) tcell.Color {
	return tcell.ColorMediumSpringGreen
}

func rbacColorer(ns string, r render.RowEvent) tcell.Color {
	return ui.DefaultColorer(ns, r)
}

func checkReadyCol(readyCol, statusCol string, c tcell.Color) tcell.Color {
	if statusCol == "Completed" {
		return c
	}

	tokens := strings.Split(readyCol, "/")
	if len(tokens) == 2 && (tokens[0] == "0" || tokens[0] != tokens[1]) {
		return ui.ErrColor
	}
	return c
}

func podColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	readyCol := 2
	if len(ns) != 0 {
		readyCol = 1
	}
	statusCol := readyCol + 1

	ready, status := strings.TrimSpace(r.Row.Fields[readyCol]), strings.TrimSpace(r.Row.Fields[statusCol])
	c = checkReadyCol(ready, status, c)

	switch status {
	case "ContainerCreating", "PodInitializing":
		return ui.AddColor
	case resource.Initialized:
		return ui.HighlightColor
	case resource.Completed:
		return ui.CompletedColor
	case resource.Running:
	case resource.Terminating:
		return ui.KillColor
	default:
		return ui.ErrColor
	}

	return c
}

func containerColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	readyCol := 2
	if strings.TrimSpace(r.Row.Fields[readyCol]) == "false" {
		c = ui.ErrColor
	}

	stateCol := readyCol + 1
	switch strings.TrimSpace(r.Row.Fields[stateCol]) {
	case "ContainerCreating", "PodInitializing":
		return ui.AddColor
	case resource.Terminating, resource.Initialized:
		return ui.HighlightColor
	case resource.Completed:
		return ui.CompletedColor
	case resource.Running:
	default:
		c = ui.ErrColor
	}

	return c
}

func ctxColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	if strings.Contains(strings.TrimSpace(r.Row.Fields[0]), "*") {
		c = ui.HighlightColor
	}

	return c
}

func pvColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	status := strings.TrimSpace(r.Row.Fields[4])
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

func pvcColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}

	if strings.TrimSpace(r.Row.Fields[markCol]) != "Bound" {
		c = ui.ErrColor
	}

	return c
}

func pdbColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	markCol := 5
	if ns != resource.AllNamespaces {
		markCol = 4
	}
	if strings.TrimSpace(r.Row.Fields[markCol]) != strings.TrimSpace(r.Row.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func dpColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Row.Fields[markCol]) != strings.TrimSpace(r.Row.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func stsColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Row.Fields[markCol]) != strings.TrimSpace(r.Row.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func rsColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	markCol := 2
	if ns != resource.AllNamespaces {
		markCol = 1
	}
	if strings.TrimSpace(r.Row.Fields[markCol]) != strings.TrimSpace(r.Row.Fields[markCol+1]) {
		return ui.ErrColor
	}

	return ui.StdColor
}

func evColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)

	markCol := 3
	if ns != resource.AllNamespaces {
		markCol = 2
	}

	switch strings.TrimSpace(r.Row.Fields[markCol]) {
	case "Failed":
		c = ui.ErrColor
	case "Killing":
		c = ui.KillColor
	}

	return c
}

func nsColorer(ns string, r render.RowEvent) tcell.Color {
	c := ui.DefaultColorer(ns, r)
	if r.Kind == render.EventAdd || r.Kind == render.EventUpdate {
		return c
	}

	switch strings.TrimSpace(r.Row.Fields[1]) {
	case "Inactive", resource.Terminating:
		c = ui.ErrColor
	}

	if strings.Contains(strings.TrimSpace(r.Row.Fields[0]), "*") {
		c = ui.HighlightColor
	}

	return c
}
