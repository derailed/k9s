// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/rs/zerolog/log"
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
func NewCronJob(gvr client.GVR) ResourceViewer {
	c := CronJob{ResourceViewer: NewVulnerabilityExtender(
		NewOwnerExtender(NewBrowser(gvr)),
	)}
	c.AddBindKeysFn(c.bindKeys)
	c.GetTable().SetEnterFn(c.showJobs)

	return &c
}

func (c *CronJob) showJobs(app *App, model ui.Tabular, gvr client.GVR, path string) {
	log.Debug().Msgf("Showing Jobs %q:%q -- %q", model.GetNamespace(), gvr, path)
	o, err := app.factory.Get(gvr.String(), path, true, labels.Everything())
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

	v := NewJob(client.NewGVR("batch/v1/jobs"))
	v.SetContextFn(jobCtx(path, string(cj.UID)))
	if err := app.inject(v, false); err != nil {
		app.Flash().Err(err)
	}
}

func jobCtx(path, uid string) ContextFunc {
	return func(ctx context.Context) context.Context {
		ctx = context.WithValue(ctx, internal.KeyPath, path)
		return context.WithValue(ctx, internal.KeyUID, uid)
	}
}

func (c *CronJob) bindKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyT:      ui.NewKeyAction("Trigger", c.triggerCmd, true),
		ui.KeyS:      ui.NewKeyAction("Suspend/Resume", c.toggleSuspendCmd, true),
		ui.KeyShiftL: ui.NewKeyAction("Sort LastScheduled", c.GetTable().SortColCmd(lastScheduledCol, true), false),
	})
}

func (c *CronJob) triggerCmd(evt *tcell.EventKey) *tcell.EventKey {
	fqn := c.GetTable().GetSelectedItem()
	if fqn == "" {
		return evt
	}

	msg := fmt.Sprintf("Trigger Cronjob %s?", fqn)
	dialog.ShowConfirm(c.App().Styles.Dialog(), c.App().Content.Pages, "Confirm Job Trigger", msg, func() {
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

		if err := runner.Run(fqn); err != nil {
			c.App().Flash().Errf("Cronjob trigger failed %v", err)
			return
		}
		c.App().Flash().Infof("Triggering Job %s %s", c.GVR(), fqn)
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

	dialog.ShowConfirm(c.App().Styles.Dialog(), c.App().Content.Pages, title, sel, func() {
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
