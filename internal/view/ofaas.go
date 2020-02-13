package view

import (
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
)

// OpenFaas represents an OpenFaaS viewer.
type OpenFaas struct {
	ResourceViewer
}

// NewOpenFaas returns a new viewer.
func NewOpenFaas(gvr client.GVR) ResourceViewer {
	o := OpenFaas{ResourceViewer: NewBrowser(gvr)}
	o.SetBindKeysFn(o.bindKeys)
	o.GetTable().SetEnterFn(o.showPods)
	o.GetTable().SetColorerFn(render.OpenFaas{}.ColorerFunc())

	return &o
}

func (o *OpenFaas) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", o.GetTable().SortColCmd(2, true), false),
		ui.KeyShiftI: ui.NewKeyAction("Sort Invocations", o.GetTable().SortColCmd(4, false), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Replicas", o.GetTable().SortColCmd(5, false), false),
		ui.KeyShiftV: ui.NewKeyAction("Sort Available", o.GetTable().SortColCmd(6, false), false),
	})
}

func (o *OpenFaas) showPods(a *App, _ ui.Tabular, _, path string) {
	labels := o.GetTable().GetSelectedCell(4)
	sels := make(map[string]string)

	tokens := strings.Split(labels, ",")
	for _, t := range tokens {
		s := strings.Split(t, "=")
		sels[s[0]] = s[1]
	}

	showPodsWithLabels(a, path, sels)
}
