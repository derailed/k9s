// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"encoding/json"
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
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
)

const (
	suspendDialogKey     = "suspend"
	triggerDialogKey     = "trigger"
	lastScheduledCol     = "LAST_SCHEDULE"
	defaultSuspendStatus = "true"
)

// CronJob represents a cronjob viewer.
type CronJob struct {
	ResourceViewer
}

type triggerContainerFields struct {
	name               string
	command, args, env *tview.InputField
}

type triggerSection struct {
	*tview.Box
	title      string
	color      tcell.Color
	finished   func(tcell.Key)
	navigation *tcell.Key
}

func newTriggerSection(title string, navigation *tcell.Key, color tcell.Color) *triggerSection {
	return &triggerSection{Box: tview.NewBox(), title: title, color: color, navigation: navigation}
}

func (s *triggerSection) Draw(screen tcell.Screen) {
	x, y, width, _ := s.GetRect()
	tview.Print(screen, s.title, x, y, width, tview.AlignCenter, s.color)
}

func (s *triggerSection) Focus(func(tview.Primitive)) {
	if s.finished == nil {
		return
	}
	key := tcell.KeyTab
	if *s.navigation == tcell.KeyBacktab {
		key = tcell.KeyBacktab
	}
	s.finished(key)
}

func (*triggerSection) GetLabel() string { return "" }

func (s *triggerSection) SetFormAttributes(_ int, _ tcell.Color, backgroundColor, _, _ tcell.Color) tview.FormItem {
	s.SetBackgroundColor(backgroundColor)
	return s
}

func (*triggerSection) GetFieldWidth() int { return 0 }

func (s *triggerSection) SetFinishedFunc(finished func(tcell.Key)) tview.FormItem {
	s.finished = finished
	return s
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
		ui.KeyT:      ui.NewKeyAction("Trigger", c.triggerCmd, true),
		ui.KeyShiftT: ui.NewKeyAction("Trigger With Edit", c.editableTriggerCmd, true),
		ui.KeyS:      ui.NewKeyAction("Suspend/Resume", c.toggleSuspendCmd, true),
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

func (c *CronJob) editableTriggerCmd(evt *tcell.EventKey) *tcell.EventKey {
	fqn := c.GetTable().GetSelectedItem()
	if fqn == "" {
		return evt
	}

	c.Stop()
	defer c.Start()
	c.showTriggerDialog(fqn)
	return nil
}

func (c *CronJob) showTriggerDialog(fqn string) {
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
	instance, err := cronJob.GetInstance(fqn)
	if err != nil {
		c.App().Flash().Err(err)
		return
	}

	styles := c.App().Styles.Dialog()
	form := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		SetFieldBackgroundColor(styles.BgColor.Color())

	containers := make([]triggerContainerFields, 0, len(instance.Spec.JobTemplate.Spec.Template.Spec.InitContainers)+len(instance.Spec.JobTemplate.Spec.Template.Spec.Containers))
	navigation := tcell.KeyTab
	trackNavigation := func(key tcell.Key) {
		navigation = key
	}
	addContainer := func(container v1.Container, init bool) {
		if len(containers) > 0 {
			form.AddFormItem(newTriggerSection("", &navigation, styles.FgColor.Color()))
		}
		form.AddFormItem(newTriggerSection(containerSectionTitle(container.Name, init), &navigation, styles.FgColor.Color()))
		command := tview.NewInputField().SetLabel("Command:").SetText(stringSliceJSON(container.Command)).SetDoneFunc(trackNavigation)
		args := tview.NewInputField().SetLabel("Args:").SetText(stringSliceJSON(container.Args)).SetDoneFunc(trackNavigation)
		env := tview.NewInputField().SetLabel("Env:").SetText(envSliceJSON(container.Env)).SetDoneFunc(trackNavigation)
		form.AddFormItem(command).AddFormItem(args).AddFormItem(env)
		containers = append(containers, triggerContainerFields{
			name:    container.Name,
			command: command,
			args:    args,
			env:     env,
		})
	}
	for _, container := range instance.Spec.JobTemplate.Spec.Template.Spec.InitContainers {
		addContainer(container, true)
	}
	for _, container := range instance.Spec.JobTemplate.Spec.Template.Spec.Containers {
		addContainer(container, false)
	}

	form.AddButton("Cancel", c.dismissTriggerDialog)
	form.AddButton("OK", func() {
		overrides, err := triggerOverrides(containers)
		if err != nil {
			c.App().Flash().Err(err)
			return
		}
		c.dismissTriggerDialog()
		if err := cronJob.RunWithOverrides(fqn, overrides); err != nil {
			c.App().Flash().Errf("CronJob trigger failed for %s: %v", fqn, err)
			return
		}
		c.App().Flash().Infof("Triggered Job %s %s", c.GVR(), fqn)
	})
	for i := range form.GetButtonCount() {
		form.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	modal := tview.NewModalForm("<Confirm Job Trigger>", form)
	modal.SetText("Edit Containers:")
	modal.SetTextColor(styles.FgColor.Color())
	modal.SetDoneFunc(func(int, string) {
		c.dismissTriggerDialog()
	})
	c.App().Content.AddPage(triggerDialogKey, modal, false, false)
	c.App().Content.ShowPage(triggerDialogKey)
	c.App().SetFocus(modal)
}

func (c *CronJob) dismissTriggerDialog() {
	c.App().Content.RemovePage(triggerDialogKey)
	c.App().SetFocus(c.GetTable())
}

func containerSectionTitle(name string, init bool) string {
	title := "[" + name + "]"
	if init {
		title += " [init]"
	}
	return title
}

func stringSliceJSON(values []string) string {
	if values == nil {
		values = []string{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func envSliceJSON(values []v1.EnvVar) string {
	if values == nil {
		values = []v1.EnvVar{}
	}
	data, _ := json.Marshal(values)
	return string(data)
}

func triggerOverrides(containers []triggerContainerFields) (map[string]dao.ContainerOverride, error) {
	overrides := make(map[string]dao.ContainerOverride, len(containers))
	for _, container := range containers {
		var command, args []string
		var env []v1.EnvVar
		if err := json.Unmarshal([]byte(container.command.GetText()), &command); err != nil || command == nil {
			return nil, fmt.Errorf("invalid command for container %q", container.name)
		}
		if err := json.Unmarshal([]byte(container.args.GetText()), &args); err != nil || args == nil {
			return nil, fmt.Errorf("invalid arguments for container %q", container.name)
		}
		if err := json.Unmarshal([]byte(container.env.GetText()), &env); err != nil || env == nil {
			return nil, fmt.Errorf("invalid env variables for container %q", container.name)
		}
		overrides[container.name] = dao.ContainerOverride{Command: command, Args: args, Env: env}
	}
	return overrides, nil
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
