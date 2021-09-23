package view

// BOZO!! revamp with latest...
// import (
// 	"strings"

// 	"github.com/derailed/k9s/internal/client"
// 	"github.com/derailed/k9s/internal/render"
// 	"github.com/derailed/k9s/internal/ui"
// )

// // OpenFaas represents an OpenFaaS viewer.
// type OpenFaas struct {
// 	ResourceViewer
// }

// // NewOpenFaas returns a new viewer.
// func NewOpenFaas(gvr client.GVR) ResourceViewer {
// 	o := OpenFaas{ResourceViewer: NewBrowser(gvr)}
// 	o.AddBindKeysFn(o.bindKeys)
// 	o.GetTable().SetEnterFn(o.showPods)
// 	o.GetTable().SetColorerFn(render.OpenFaas{}.ColorerFunc())

// 	return &o
// }

// func (o *OpenFaas) bindKeys(aa ui.KeyActions) {
// 	aa.Add(ui.KeyActions{
// 		ui.KeyShiftS: ui.NewKeyAction("Sort Status", o.GetTable().SortColCmd(statusCol, true), false),
// 		ui.KeyShiftI: ui.NewKeyAction("Sort Invocations", o.GetTable().SortColCmd("INVOCATIONS", false), false),
// 		ui.KeyShiftR: ui.NewKeyAction("Sort Replicas", o.GetTable().SortColCmd("REPLICAS", false), false),
// 		ui.KeyShiftL: ui.NewKeyAction("Sort Available", o.GetTable().SortColCmd(availCol, false), false),
// 	})
// }

// func (o *OpenFaas) showPods(a *App, _ ui.Tabular, _, path string) {
// 	labels := o.GetTable().GetSelectedCell(o.GetTable().NameColIndex() + 3)
// 	sels := make(map[string]string)

// 	tokens := strings.Split(labels, ",")
// 	for _, t := range tokens {
// 		s := strings.Split(t, "=")
// 		sels[s[0]] = s[1]
// 	}

// 	showPodsWithLabels(a, path, sels)
// }
