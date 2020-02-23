package view

import (
	"context"
	"fmt"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	batchv1beta1 "k8s.io/api/batch/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// CronJob represents a cronjob viewer.
type CronJob struct {
	ResourceViewer
}

// NewCronJob returns a new viewer.
func NewCronJob(gvr client.GVR) ResourceViewer {
	c := CronJob{ResourceViewer: NewBrowser(gvr)}
	c.SetBindKeysFn(c.bindKeys)
	c.GetTable().SetEnterFn(c.showJobs)
	c.GetTable().SetColorerFn(render.CronJob{}.ColorerFunc())

	return &c
}

func (c *CronJob) showJobs(app *App, model ui.Tabular, gvr, path string) {
	log.Debug().Msgf("Showing Jobs %q:%q -- %q", model.GetNamespace(), gvr, path)
	o, err := app.factory.Get(gvr, path, true, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var cj batchv1beta1.CronJob
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &cj)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	v := NewJob(client.NewGVR("batch/v1/jobs"))
	v.SetContextFn(jobCtx(path, string(cj.UID)))
	if err := app.inject(v); err != nil {
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
		tcell.KeyCtrlT: ui.NewKeyAction("Trigger", c.trigger, true),
	})
}

func (c *CronJob) trigger(evt *tcell.EventKey) *tcell.EventKey {
	sel := c.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	res, err := dao.AccessorFor(c.App().factory, c.GVR())
	if err != nil {
		return nil
	}
	runner, ok := res.(dao.Runnable)
	if !ok {
		c.App().Flash().Err(fmt.Errorf("expecting a jobrunner resource for %q", c.GVR()))
		return nil
	}

	if err := runner.Run(sel); err != nil {
		c.App().Flash().Errf("Cronjob trigger failed %v", err)
		return evt
	}
	c.App().Flash().Infof("Triggering Job %s %s", c.GVR(), sel)

	return nil
}
