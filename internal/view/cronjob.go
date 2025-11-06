// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	suspendDialogKey     = "suspend"
	lastScheduledCol     = "LAST_SCHEDULE"
	defaultSuspendStatus = "true"
)

// CronJob represents a cronjob viewer.
type CronJob struct {
	ResourceViewer
}

// NewCronJob returns a new viewer.
func NewCronJob(gvr *client.GVR) ResourceViewer {
	c := CronJob{ResourceViewer: NewVulnerabilityExtender(
		NewOwnerExtender(NewBrowser(gvr)),
	)}
	c.AddBindKeysFn(c.bindKeys)
	c.GetTable().SetEnterFn(c.showJobs)

	return &c
}

func (*CronJob) showJobs(app *App, _ ui.Tabular, gvr *client.GVR, fqn string) {
	slog.Debug("Showing Jobs", slogs.GVR, gvr, slogs.FQN, fqn)
	o, err := app.factory.Get(gvr, fqn, true, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var cj batchv1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	ns, _ := client.Namespaced(fqn)
	if err := app.Config.SetActiveNamespace(ns); err != nil {
		slog.Error("Unable to set active namespace during show pods", slogs.Error, err)
	}
	v := NewJob(client.JobGVR)
	v.SetContextFn(jobCtx(fqn, string(cj.UID)))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func jobCtx(fqn, uid string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, fqn)
		return context.WithValue(ctx, internal.KeyUID, uid)
	}
}

func (c *CronJob) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyT: ui.NewKeyAction("Trigger", c.triggerCmd, true),
		ui.KeyS: ui.NewKeyAction("Suspend/Resume", c.toggleSuspendCmd, true),
	})
}

func (c *CronJob) triggerCmd(evt *tcell.EventKey) *tcell.EventKey {
	fqns := c.GetTable().GetSelectedItems()
	if len(fqns) == 0 {
		return evt
	}
	msg := fmt.Sprintf("Trigger CronJob: %s?", fqns[0])
	if len(fqns) > 1 {
		msg = fmt.Sprintf("Trigger %d CronJobs?", len(fqns))
	}
	d := c.App().Styles.Dialog()
	dialog.ShowConfirm(&d, c.App().Content.Pages, "Confirm Job Trigger", msg, func() {
		res, err := dao.AccessorFor(c.App().factory, c.GVR())
		if err != nil {
			c.App().Flash().Err(fmt.Errorf("no accessor for %q", c.GVR()))
			return
		}
		runner, ok := res.(dao.Runnable)
		if !ok {
			c.App().Flash().Err(fmt.Errorf("expecting a job runner resource for %q", c.GVR()))
			return
		}

		for _, fqn := range fqns {
			if err := runner.Run(fqn); err != nil {
				c.App().Flash().Errf("CronJob trigger failed for %s: %v", fqn, err)
			} else {
				c.App().Flash().Infof("Triggered Job %s %s", c.GVR(), fqn)
			}
		}
	}, func() {})

	return nil
}

func (c *CronJob) toggleSuspendCmd(evt *tcell.EventKey) *tcell.EventKey {
	table := c.GetTable()
	sel := table.GetSelectedItem()

	if sel == "" {
		return evt
	}

	cell := table.GetCell(c.GetTable().GetSelectedRowIndex(), c.GetTable().NameColIndex()+2)

	if cell == nil {
		c.App().Flash().Errf("Unable to assert current status")
		return nil
	}

	c.Stop()
	defer c.Start()

	c.showSuspendDialog(cell, sel)

	return nil
}

func (c *CronJob) showSuspendDialog(cell *tview.TableCell, sel string) {
	title := "Suspend"

	if strings.TrimSpace(cell.Text) == defaultSuspendStatus {
		title = "Resume"
	}

	d := c.App().Styles.Dialog()
	dialog.ShowConfirm(&d, c.App().Content.Pages, title, sel, func() {
		ctx, cancel := context.WithTimeout(context.Background(), c.App().Conn().Config().CallTimeout())
		defer cancel()

		res, err := dao.AccessorFor(c.App().factory, c.GVR())
		if err != nil {
			c.App().Flash().Err(fmt.Errorf("no accessor for %q", c.GVR()))
			return
		}

		cronJob, ok := res.(*dao.CronJob)
		if !ok {
			c.App().Flash().Errf("expecting a cron job for %q", c.GVR())
			return
		}

		if err := cronJob.ToggleSuspend(ctx, sel); err != nil {
			c.App().Flash().Errf("Cronjob %s failed for %v", strings.ToLower(title), err)
			return
		}
	}, func() {})
}
