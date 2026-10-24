// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	batchv1 "k8s.io/api/batch/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

// Job represents a job viewer.
type Job struct {
	ResourceViewer
}

// NewJob returns a new viewer.
func NewJob(gvr *client.GVR) ResourceViewer {
	var j Job

	j.ResourceViewer = NewVulnerabilityExtender(
		NewOwnerExtender(
			NewLogsExtender(NewBrowser(gvr), j.logOptions),
		),
	)
	j.AddBindKeysFn(j.bindKeys)
	j.GetTable().SetEnterFn(j.showPods)
	j.GetTable().SetSortCol("AGE", true)

	return &j
}

func (j *Job) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyS, ui.NewKeyAction("Suspend/Resume", j.toggleSuspendCmd, true))
}

func (j *Job) toggleSuspendCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := j.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	job, err := j.getInstance(path)
	if err != nil {
		j.App().Flash().Err(err)
		return nil
	}

	title := "Suspend"
	if job.Spec.Suspend != nil && *job.Spec.Suspend {
		title = "Resume"
	}

	d := j.App().Styles.Dialog()
	dialog.ShowConfirm(&d, j.App().Content.Pages, title, path, func() {
		ctx, cancel := context.WithTimeout(context.Background(), j.App().Conn().Config().CallTimeout())
		defer cancel()

		var job dao.Job
		job.Init(j.App().factory, client.JobGVR)
		if err := job.ToggleSuspend(ctx, path); err != nil {
			j.App().Flash().Errf("Job %s failed: %v", strings.ToLower(title), err)
		}
	}, func() {})

	return nil
}

func (*Job) showPods(app *App, _ ui.Tabular, gvr *client.GVR, path string) {
	o, err := app.factory.Get(gvr, path, true, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var job batchv1.Job
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &job)
	if err != nil {
		app.Flash().Err(err)
		return
	}

	showPodsFromSelector(app, path, job.Spec.Selector)
}

func (j *Job) logOptions(prev bool) (*dao.LogOptions, error) {
	path := j.GetTable().GetSelectedItem()
	if path == "" {
		return nil, errors.New("you must provide a selection")
	}
	job, err := j.getInstance(path)
	if err != nil {
		return nil, err
	}

	return podLogOptions(j.App(), path, prev, &job.ObjectMeta, &job.Spec.Template.Spec), nil
}

func (j *Job) getInstance(fqn string) (*batchv1.Job, error) {
	var job dao.Job
	job.Init(j.App().factory, client.JobGVR)

	return job.GetInstance(fqn)
}
