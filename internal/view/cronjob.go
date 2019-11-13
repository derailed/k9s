package view

import (
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// CronJob presents a cronjob viewer.
type CronJob struct {
	*Resource
}

// NewCronJob returns a new viewer.
func NewCronJob(title, gvr string, list resource.List) ResourceViewer {
	c := CronJob{
		Resource: NewResource(title, gvr, list),
	}
	c.extraActionsFn = c.extraActions

	return &c
}

func (c *CronJob) trigger(evt *tcell.EventKey) *tcell.EventKey {
	if !c.masterPage().RowSelected() {
		return evt
	}

	sel := c.masterPage().GetSelectedItem()
	if err := c.list.Resource().(resource.Runner).Run(sel); err != nil {
		c.app.Flash().Errf("Cronjob trigger failed %v", err)
		return evt
	}
	c.app.Flash().Infof("Triggering %s %s", c.list.GetName(), sel)

	return nil
}

func (c *CronJob) extraActions(aa ui.KeyActions) {
	aa[tcell.KeyCtrlT] = ui.NewKeyAction("Trigger", c.trigger, true)
}
