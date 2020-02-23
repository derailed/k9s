package view

import (
	"context"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
)

// Chart represents a helm chart view.
type Chart struct {
	ResourceViewer
}

// NewChart returns a new alias view.
func NewChart(gvr client.GVR) ResourceViewer {
	c := Chart{
		ResourceViewer: NewBrowser(gvr),
	}
	c.GetTable().SetColorerFn(render.Chart{}.ColorerFunc())
	c.GetTable().SetBorderFocusColor(tcell.ColorMediumSpringGreen)
	c.GetTable().SetSelectedStyle(tcell.ColorWhite, tcell.ColorMediumSpringGreen, tcell.AttrNone)
	c.SetBindKeysFn(c.bindKeys)
	c.SetContextFn(c.chartContext)

	return &c
}

func (c *Chart) chartContext(ctx context.Context) context.Context {
	return ctx
}

func (c *Chart) bindKeys(aa ui.KeyActions) {
	aa.Delete(ui.KeyShiftA, ui.KeyShiftN, tcell.KeyCtrlS, tcell.KeyCtrlSpace, ui.KeySpace)
	aa.Add(ui.KeyActions{
		ui.KeyB:      ui.NewKeyAction("Blee", c.bleeCmd, true),
		ui.KeyShiftN: ui.NewKeyAction("Sort Name", c.GetTable().SortColCmd(nameCol, true), false),
		ui.KeyShiftS: ui.NewKeyAction("Sort Status", c.GetTable().SortColCmd(statusCol, true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", c.GetTable().SortColCmd(ageCol, true), false),
	})
}

func (c *Chart) bleeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := c.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}
	log.Debug().Msgf("BLEE CMD %q", path)
	return nil
}
