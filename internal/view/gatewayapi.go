// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// GatewayAPI presents a Gateway API dashboard viewer.
type GatewayAPI struct {
	ResourceViewer
}

// NewGatewayAPI returns a new Gateway API viewer.
func NewGatewayAPI(gvr *client.GVR) ResourceViewer {
	g := GatewayAPI{
		ResourceViewer: NewBrowser(gvr),
	}
	g.GetTable().SetEnterFn(g.showRes)
	g.AddBindKeysFn(g.bindKeys)
	g.GetTable().SetSortCol("RESOURCE_TYPE", true)

	return &g
}

func (g *GatewayAPI) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftK: ui.NewKeyAction("Sort Resource Type", g.GetTable().SortColCmd("RESOURCE_TYPE", true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", g.GetTable().SortColCmd("AGE", true), false),
		ui.KeyY:      ui.NewKeyAction("YAML", g.yamlCmd, true),
		ui.KeyD:      ui.NewKeyAction("Describe", g.describeCmd, true),
	})
}

func (*GatewayAPI) showRes(app *App, _ ui.Tabular, _ *client.GVR, path string) {
	gvr, fqn, ok := parsePath(path)
	if !ok {
		app.Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return
	}
	app.gotoResource(gvr.String(), fqn, false, true)
}

func (g *GatewayAPI) yamlCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := g.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		g.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	v := NewLiveView(g.App(), yamlAction, model.NewYAML(gvr, fqn))
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}

	return nil
}

func (g *GatewayAPI) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := g.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		g.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	describeResource(g.App(), nil, gvr, fqn)

	return nil
}