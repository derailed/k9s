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
	c := CronJob{ResourceViewer: NewBrowser(gvr)}
	c.AddBindKeysFn(c.bindKeys)
	c.GetTable().SetEnterFn(c.showJobs)

	return &c
}

func (c *CronJob) showJobs(app *App, model ui.Tabular, gvr, path string) {
	log.Debug().Msgf("Showing Jobs %q:%q -- %q", model.GetNamespace(), gvr, path)
	o, err := app.factory.Get(gvr, path, true, labels.Everything())
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

func (c *CronJob) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
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
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	c.Stop()
	defer c.Start()
	c.showSuspendDialog(sel)

	return nil
}

func (c *CronJob) showSuspendDialog(sel string) {
	cell := c.GetTable().GetCell(c.GetTable().GetSelectedRowIndex(), c.GetTable().NameColIndex()+2)
	if cell == nil {
		c.App().Flash().Errf("Unable to assert current status")
		return
	}
	suspended := strings.TrimSpace(cell.Text) == defaultSuspendStatus
	title := "Suspend"
	if suspended {
		title = "Resume"
	}

	confirm := tview.NewModalForm(fmt.Sprintf("<%s>", title), c.makeSuspendForm(sel, !suspended))
	confirm.SetText(fmt.Sprintf("%s CronJob %s?", title, sel))
	confirm.SetDoneFunc(func(int, string) {
		c.dismissDialog()
	})
	c.App().Content.AddPage(suspendDialogKey, confirm, false, false)
	c.App().Content.ShowPage(suspendDialogKey)
}

func (c *CronJob) makeSuspendForm(sel string, suspend bool) *tview.Form {
	f := c.makeStyledForm()
	action := "suspended"
	if !suspend {
		action = "resumed"
	}

	f.AddButton("Cancel", func() {
		c.dismissDialog()
	})
	f.AddButton("OK", func() {
		defer c.dismissDialog()

		ctx, cancel := context.WithTimeout(context.Background(), c.App().Conn().Config().CallTimeout())
		defer cancel()
		if err := c.toggleSuspend(ctx, sel); err != nil {
			log.Error().Err(err).Msgf("CronJob %s %s failed", sel, action)
			c.App().Flash().Err(err)
		} else {
			c.App().Flash().Infof("CronJob %s %s successfully!", sel, action)
		}
	})

	return f
}

func (c *CronJob) toggleSuspend(ctx context.Context, path string) error {
	res, err := dao.AccessorFor(c.App().factory, c.GVR())
	if err != nil {
		return err
	}
	cronJob, ok := res.(*dao.CronJob)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", c.GVR())
	}

	return cronJob.ToggleSuspend(ctx, path)
}

func (c *CronJob) makeStyledForm() *tview.Form {
	f := tview.NewForm()
	f.SetItemPadding(0)
	f.SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(tview.Styles.PrimitiveBackgroundColor).
		SetButtonTextColor(tview.Styles.PrimaryTextColor).
		SetLabelColor(tcell.ColorAqua).
		SetFieldTextColor(tcell.ColorOrange)

	return f
}

func (c *CronJob) dismissDialog() {
	c.App().Content.RemovePage(suspendDialogKey)
}
