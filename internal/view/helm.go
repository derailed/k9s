package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell/v2"
	"github.com/rs/zerolog/log"
)

// Helm represents a helm chart view.
type Helm struct {
	ResourceViewer

	Values *model.Values
}

// NewHelm returns a new alias view.
func NewHelm(gvr client.GVR) ResourceViewer {
	c := Helm{
		ResourceViewer: NewBrowser(gvr),
	}
	c.GetTable().SetColorerFn(render.Helm{}.ColorerFunc())
	c.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	c.GetTable().SetSelectedStyle(tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorMediumSpringGreen).Attributes(tcell.AttrNone))
	c.AddBindKeysFn(c.bindKeys)
	c.SetContextFn(c.chartContext)

	return &c
}

func (c *Helm) chartContext(ctx context.Context) context.Context {
	return ctx
}

func (c *Helm) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyShiftN: ui.NewKeyAction("Sort Name", c.GetTable().SortColCmd(nameCol, true), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", c.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", c.GetTable().SortColCmd(ageCol, true), false),
		ui.KeyV:      ui.NewKeyAction("Values", c.getValsCmd(), true),
	})
}

func (c *Helm) getValsCmd() func(evt *tcell.EventKey) *tcell.EventKey {
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
		if err := v.app.inject(v); err != nil {
			v.app.Flash().Err(err)
		}
		return nil
	}
}

func (c *Helm) toggleValuesCmd(evt *tcell.EventKey) *tcell.EventKey {
	c.Values.ToggleValues()
	if err := c.Values.Refresh(c.defaultCtx()); err != nil {
		log.Error().Err(err).Msgf("helm refresh failed")
		return nil
	}
	c.App().Flash().Infof("Values toggled")
	return nil
}

func (c *Helm) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, c.App().factory)
}
