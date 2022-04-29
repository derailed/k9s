package view

import (
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
)

// Dashboard represents a dashboard view.
type Dashboard struct {
	ResourceViewer
}

// NewDashboard returns a new view.
func NewDashboard(gvr client.GVR) ResourceViewer {
	dash := Dashboard{ResourceViewer: NewBrowser(gvr)}

	table := dash.GetTable()
	table.SetColorerFn(render.Dashboard{}.ColorerFunc())
	table.SetEnterFn(dash.EnterFunc)
	table.SetSortCol("RESOURCE", true)

	return &dash
}

func (dash Dashboard) EnterFunc(app *App, model ui.Tabular, gvr, path string) {
	app.command.run(path, "", false)
}
