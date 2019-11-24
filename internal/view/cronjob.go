package view

import (
	"context"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

// CronJob presents a cronjob viewer.
type CronJob struct {
	ResourceViewer
}

// NewCronJob returns a new viewer.
func NewCronJob(title, gvr string, list resource.List) ResourceViewer {
	return &CronJob{
		ResourceViewer: NewResource(title, gvr, list).(ResourceViewer),
	}
}

func (c *CronJob) Init(ctx context.Context) error {
	if err := c.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	c.bindKeys()

	return nil
}

func (c *CronJob) bindKeys() {
	c.Actions().Add(ui.KeyActions{
		tcell.KeyCtrlT: ui.NewKeyAction("Trigger", c.trigger, true),
	})
}

func (c *CronJob) trigger(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	if err := c.List().Resource().(resource.Runner).Run(sel); err != nil {
		c.App().Flash().Errf("Cronjob trigger failed %v", err)
		return evt
	}
	c.App().Flash().Infof("Triggering %s %s", c.List().GetName(), sel)

	return nil
}
