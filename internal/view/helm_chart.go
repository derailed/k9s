// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// HelmChart represents a helm chart view.
type HelmChart struct {
	ResourceViewer
}

// NewHelmChart returns a new helm-chart view.
func NewHelmChart(gvr client.GVR) ResourceViewer {
	c := HelmChart{
		ResourceViewer: NewValueExtender(NewBrowser(gvr)),
	}
	c.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	c.GetTable().SetSelectedStyle(tcell.StyleDefault.
		Foreground(tcell.ColorWhite).
		Background(tcell.ColorMediumSpringGreen).Attributes(tcell.AttrNone))
	c.AddBindKeysFn(c.bindKeys)
	c.GetTable().SetEnterFn(c.viewReleases)
	c.SetContextFn(c.chartContext)

	return &c
}

func (c *HelmChart) chartContext(ctx context.Context) context.Context {
	return ctx
}

func (c *HelmChart) bindKeys(aa *ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlS)
	aa.Bulk(ui.KeyMap{
		ui.KeyR:      ui.NewKeyAction("Releases", c.historyCmd, true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", c.GetTable().SortColCmd(statusCol, true), false),
	})
}

func (c *HelmChart) viewReleases(app *App, model ui.Tabular, _ client.GVR, path string) {
	v := NewHistory(client.NewGVR("helm-history"))
	v.SetContextFn(c.helmContext)
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func (c *HelmChart) historyCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	c.viewReleases(c.App(), c.GetTable().GetModel(), c.GVR(), path)

	return nil
}

func (c *HelmChart) helmContext(ctx context.Context) context.Context {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return ctx
	}
	ctx = context.WithValue(ctx, internal.KeyFQN, path)

	return context.WithValue(ctx, internal.KeyPath, path)
}
