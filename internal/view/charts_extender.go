// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"errors"

	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

var errNoMetrics = errors.New("metrics-server unavailable")

// ChartsExtender adds a container metrics chart action to a resource viewer.
type ChartsExtender struct {
	ResourceViewer
}

// NewChartsExtender returns a new extender.
func NewChartsExtender(v ResourceViewer) ResourceViewer {
	c := ChartsExtender{ResourceViewer: v}
	c.AddBindKeysFn(c.bindKeys)

	return &c
}

func (c *ChartsExtender) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyShiftG: ui.NewKeyAction("Charts", c.chartsCmd, true),
	})
}

func (c *ChartsExtender) chartsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !c.App().Conn().HasMetrics() {
		c.App().Flash().Err(errNoMetrics)
		return nil
	}

	container := c.GetTable().GetSelectedItem()
	if container == "" {
		return evt
	}

	if err := c.App().inject(NewContainerCharts(c.GetTable().Path, container), false); err != nil {
		c.App().Flash().Err(err)
	}

	return nil
}
