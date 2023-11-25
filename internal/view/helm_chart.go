// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// HelmChart represents a helm chart view.
type HelmChart struct {
	ResourceViewer

	Values *model.Values
}

// NewHelm returns a new alias view.
func NewHelmChart(gvr client.GVR) ResourceViewer {
	c := HelmChart{
		ResourceViewer: NewBrowser(gvr),
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

func (c *HelmChart) bindKeys(aa ui.KeyActions) {
	aa.Delete(tcell.KeyCtrlS)
	aa.Add(ui.KeyActions{
		ui.KeyR:      ui.NewKeyAction("Releases", c.historyCmd, true),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", c.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyV:      ui.NewKeyAction("Values", c.getValsCmd(), true),
	})
}

func (c *HelmChart) viewReleases(app *App, model ui.Tabular, _, path string) {
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
	c.viewReleases(c.App(), c.GetTable().GetModel(), c.GVR().String(), path)

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

func (c *HelmChart) getValsCmd() func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		path := c.GetTable().GetSelectedItem()
		if path == "" {
			return evt
		}
		c.Values = model.NewValues(c.GVR(), path)
		v := NewLiveView(c.App(), "Values", c.Values)
		v.actions.Add(ui.KeyActions{
			ui.KeyV: ui.NewKeyAction("Toggle All Values", c.toggleValuesCmd, true),
		})
		if err := v.app.inject(v, false); err != nil {
			v.app.Flash().Err(err)
		}
		return nil
	}
}

func (c *HelmChart) toggleValuesCmd(evt *tcell.EventKey) *tcell.EventKey {
	c.Values.ToggleValues()
	if err := c.Values.Refresh(c.defaultCtx()); err != nil {
		log.Error().Err(err).Msgf("helm refresh failed")
		return nil
	}
	c.App().Flash().Infof("Values toggled")
	return nil
}

func (c *HelmChart) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, c.App().factory)
}
