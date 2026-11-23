// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/slogs"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type workloadTarget struct {
	gvr *client.GVR
	fqn string
}

// Workload presents a workload viewer.
type Workload struct {
	ResourceViewer
}

// NewWorkload returns a new viewer.
func NewWorkload(gvr *client.GVR) ResourceViewer {
	w := Workload{
		ResourceViewer: NewBrowser(gvr),
	}
	w.GetTable().SetEnterFn(w.showRes)
	w.AddBindKeysFn(w.bindKeys)
	w.GetTable().SetSortCol("KIND", true)

	return &w
}

func (w *Workload) bindDangerousKeys(aa *ui.KeyActions) {
	aa.Bulk(ui.KeyMap{
		ui.KeyE: ui.NewKeyActionWithOpts("Edit", w.editCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		tcell.KeyCtrlD: ui.NewKeyActionWithOpts("Delete", w.deleteCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyR: ui.NewKeyActionWithOpts("Restart", w.restartCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyS: ui.NewKeyActionWithOpts("Scale", w.scaleCmd,
			ui.ActionOpts{
				Visible:   true,
				Dangerous: true,
			}),
		ui.KeyI: ui.NewKeyActionWithOpts("Set Image", w.setImageCmd,
			ui.ActionOpts{
				Visible:   false,
				Dangerous: true,
			}),
	})
}

func (w *Workload) bindKeys(aa *ui.KeyActions) {
	if !w.App().Config.IsReadOnly() {
		w.bindDangerousKeys(aa)
	}

	aa.Bulk(ui.KeyMap{
		ui.KeyShiftK: ui.NewKeyAction("Sort Kind", w.GetTable().SortColCmd("KIND", true), false),
		ui.KeyShiftR: ui.NewKeyAction("Sort Ready", w.GetTable().SortColCmd("READY", true), false),
		ui.KeyShiftA: ui.NewKeyAction("Sort Age", w.GetTable().SortColCmd(ageCol, true), false),
		ui.KeyY:      ui.NewKeyAction(yamlAction, w.yamlCmd, true),
		ui.KeyD:      ui.NewKeyAction("Describe", w.describeCmd, true),
		ui.KeyL:      ui.NewKeyAction("Logs", w.logsCmd(false), true),
		ui.KeyP:      ui.NewKeyAction("Logs Previous", w.logsCmd(true), true),
		ui.KeyF:      ui.NewKeyAction("Show PortForward", w.showPFCmd, true),
		ui.KeyShiftF: ui.NewKeyAction("Port-Forward", w.portFwdCmd, true),
		ui.KeyZ:      ui.NewKeyAction("ReplicaSets", w.replicaSetsCmd, true),
	})
}

func parsePath(path string) (*client.GVR, string, bool) {
	tt := strings.Split(path, "|")
	if len(tt) != 3 {
		slog.Error("Unable to parse workload path", slogs.Path, path)
		return client.NoGVR, client.FQN("", ""), false
	}

	return client.NewGVR(tt[0]), client.FQN(tt[1], tt[2]), true
}

func isSupportedWorkloadGVR(gvr *client.GVR) bool {
	return gvr == client.DpGVR || gvr == client.StsGVR
}

func (w *Workload) selectedTarget() (workloadTarget, bool) {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return workloadTarget{}, false
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return workloadTarget{}, false
	}

	return workloadTarget{gvr: gvr, fqn: fqn}, true
}

func (w *Workload) selectedTargets() []workloadTarget {
	paths := w.GetTable().GetSelectedItems()
	targets := make([]workloadTarget, 0, len(paths))
	for _, path := range paths {
		gvr, fqn, ok := parsePath(path)
		if !ok || !isSupportedWorkloadGVR(gvr) {
			continue
		}
		targets = append(targets, workloadTarget{gvr: gvr, fqn: fqn})
	}

	return targets
}

func getDeployment(f dao.Factory, fqn string) (*appsv1.Deployment, error) {
	var dp dao.Deployment
	dp.Init(f, client.DpGVR)

	return dp.GetInstance(fqn)
}

func getStatefulSet(f dao.Factory, fqn string) (*appsv1.StatefulSet, error) {
	var sts dao.StatefulSet

	return sts.GetInstance(f, fqn)
}

func (*Workload) showRes(app *App, _ ui.Tabular, _ *client.GVR, path string) {
	gvr, fqn, ok := parsePath(path)
	if !ok {
		app.Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return
	}
	switch gvr {
	case client.DpGVR:
		dp, err := getDeployment(app.factory, fqn)
		if err != nil {
			app.Flash().Err(err)
			return
		}
		showPodsFromSelector(app, fqn, dp.Spec.Selector)
	case client.StsGVR:
		sts, err := getStatefulSet(app.factory, fqn)
		if err != nil {
			app.Flash().Err(err)
			return
		}
		showPodsFromSelector(app, fqn, sts.Spec.Selector)
	default:
		app.gotoResource(gvr.String(), fqn, false, true)
	}
}

func (w *Workload) deleteCmd(evt *tcell.EventKey) *tcell.EventKey {
	selections := w.GetTable().GetSelectedItems()
	if len(selections) == 0 {
		return evt
	}

	w.Stop()
	defer w.Start()
	{
		msg := fmt.Sprintf("Delete %s %s?", w.GVR().R(), selections[0])
		if len(selections) > 1 {
			msg = fmt.Sprintf("Delete %d marked %s?", len(selections), w.GVR())
		}
		w.resourceDelete(selections, msg)
	}

	return nil
}

func (w *Workload) defaultContext(gvr *client.GVR, fqn string) context.Context {
	ctx := context.WithValue(context.Background(), internal.KeyFactory, w.App().factory)
	ctx = context.WithValue(ctx, internal.KeyGVR, gvr)
	if fqn != "" {
		ctx = context.WithValue(ctx, internal.KeyPath, fqn)
	}
	if internal.IsLabelSelector(w.GetTable().CmdBuff().GetText()) {
		if sel, err := ui.ExtractLabelSelector(w.GetTable().CmdBuff().GetText()); err == nil {
			ctx = context.WithValue(ctx, internal.KeyLabels, sel)
		}
	}
	ctx = context.WithValue(ctx, internal.KeyNamespace, client.CleanseNamespace(w.App().Config.ActiveNamespace()))
	ctx = context.WithValue(ctx, internal.KeyWithMetrics, w.App().factory.Client().HasMetrics())

	return ctx
}

func (w *Workload) resourceDelete(selections []string, msg string) {
	okFn := func(propagation *metav1.DeletionPropagation, force bool) {
		w.GetTable().ShowDeleted()
		if len(selections) > 1 {
			w.App().Flash().Infof("Delete %d marked %s", len(selections), w.GVR())
		} else {
			w.App().Flash().Infof("Delete resource %s %s", w.GVR(), selections[0])
		}
		for _, sel := range selections {
			gvr, fqn, ok := parsePath(sel)
			if !ok {
				w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", sel))
				return
			}

			grace := dao.DefaultGrace
			if force {
				grace = dao.ForceGrace
			}
			if err := w.GetTable().GetModel().Delete(w.defaultContext(gvr, fqn), fqn, propagation, grace); err != nil {
				w.App().Flash().Errf("Delete failed with `%s", err)
			} else {
				w.App().factory.DeleteForwarder(sel)
			}
			w.GetTable().DeleteMark(sel)
		}
		w.GetTable().Start()
	}
	d := w.App().Styles.Dialog()
	dialog.ShowDelete(&d, w.App().Content.Pages, msg, okFn, func() {})
}

func (w *Workload) describeCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	describeResource(w.App(), nil, gvr, fqn)

	return nil
}

func (w *Workload) editCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	w.Stop()
	defer w.Start()
	if err := editRes(w.App(), gvr, fqn); err != nil {
		w.App().Flash().Err(err)
	}

	return nil
}

func (w *Workload) yamlCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := w.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}
	gvr, fqn, ok := parsePath(path)
	if !ok {
		w.App().Flash().Err(fmt.Errorf("unable to parse path: %q", path))
		return evt
	}

	v := NewLiveView(w.App(), yamlAction, model.NewYAML(gvr, fqn))
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}

	return nil
}

func (w *Workload) logsCmd(prev bool) func(*tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t, ok := w.selectedTarget()
		if !ok || !isSupportedWorkloadGVR(t.gvr) {
			return evt
		}
		w.showLogs(t, prev)

		return nil
	}
}

func (w *Workload) showLogs(t workloadTarget, prev bool) {
	ns, _ := client.Namespaced(t.fqn)
	if _, err := w.App().factory.CanForResource(ns, client.PodGVR, client.ListAccess); err != nil {
		w.App().Flash().Err(err)
		return
	}
	opts, err := w.logOptions(t, prev)
	if err != nil {
		w.App().Flash().Err(err)
		return
	}
	if err := w.App().inject(NewLog(t.gvr, opts), false); err != nil {
		w.App().Flash().Err(err)
	}
}

func (w *Workload) logOptions(t workloadTarget, prev bool) (*dao.LogOptions, error) {
	switch t.gvr {
	case client.DpGVR:
		dp, err := getDeployment(w.App().factory, t.fqn)
		if err != nil {
			return nil, err
		}
		return podLogOptions(w.App(), t.fqn, prev, &dp.ObjectMeta, &dp.Spec.Template.Spec), nil
	case client.StsGVR:
		sts, err := getStatefulSet(w.App().factory, t.fqn)
		if err != nil {
			return nil, err
		}
		return podLogOptions(w.App(), t.fqn, prev, &sts.ObjectMeta, &sts.Spec.Template.Spec), nil
	default:
		return nil, fmt.Errorf("logs not supported for %s", t.gvr)
	}
}

func (w *Workload) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	targets := w.selectedTargets()
	if len(targets) == 0 {
		return evt
	}

	w.Stop()
	defer w.Start()
	msg := fmt.Sprintf("Restart %s %s?", singularize(targets[0].gvr.R()), targets[0].fqn)
	if len(targets) > 1 {
		msg = fmt.Sprintf("Restart %d %s?", len(targets), w.GVR().R())
	}
	d := w.App().Styles.Dialog()
	opts := dialog.RestartDialogOpts{
		Title:        "Confirm Restart",
		Message:      msg,
		FieldManager: "kubectl-rollout",
		Ack: func(opts *metav1.PatchOptions) bool {
			ctx, cancel := context.WithTimeout(context.Background(), w.App().Conn().Config().CallTimeout())
			defer cancel()
			for _, t := range targets {
				if err := w.restartRollout(ctx, t, opts); err != nil {
					w.App().Flash().Err(err)
				} else {
					w.App().Flash().Infof("Restart in progress for `%s...", t.fqn)
				}
			}
			return true
		},
		Cancel: func() {},
	}
	dialog.ShowRestart(&d, w.App().Content.Pages, &opts)

	return nil
}

func (w *Workload) restartRollout(ctx context.Context, t workloadTarget, opts *metav1.PatchOptions) error {
	res, err := dao.AccessorFor(w.App().factory, t.gvr)
	if err != nil {
		return err
	}
	r, ok := res.(dao.Restartable)
	if !ok {
		return errors.New("resource is not restartable")
	}

	return r.Restart(ctx, t.fqn, opts)
}

func (w *Workload) scaleCmd(evt *tcell.EventKey) *tcell.EventKey {
	targets := w.selectedTargets()
	if len(targets) == 0 {
		return evt
	}

	w.Stop()
	defer w.Start()
	w.showScaleDialog(targets)

	return nil
}

func (w *Workload) showScaleDialog(targets []workloadTarget) {
	form, err := w.makeScaleForm(targets)
	if err != nil {
		w.App().Flash().Err(err)
		return
	}
	confirm := tview.NewModalForm("<Scale>", form)
	msg := fmt.Sprintf("Scale %s %s?", singularize(targets[0].gvr.R()), targets[0].fqn)
	if len(targets) > 1 {
		msg = fmt.Sprintf("Scale [%d] %s?", len(targets), w.GVR().R())
	}
	confirm.SetText(msg)
	confirm.SetDoneFunc(func(int, string) {
		w.App().Content.RemovePage(scaleDialogKey)
	})
	w.App().Content.AddPage(scaleDialogKey, confirm, false, false)
	w.App().Content.ShowPage(scaleDialogKey)
}

func (w *Workload) currentReplicas(t workloadTarget) (string, error) {
	switch t.gvr {
	case client.DpGVR:
		dp, err := getDeployment(w.App().factory, t.fqn)
		if err != nil {
			return "", err
		}
		if dp.Spec.Replicas == nil {
			return "1", nil
		}
		return strconv.Itoa(int(*dp.Spec.Replicas)), nil
	case client.StsGVR:
		sts, err := getStatefulSet(w.App().factory, t.fqn)
		if err != nil {
			return "", err
		}
		if sts.Spec.Replicas == nil {
			return "1", nil
		}
		return strconv.Itoa(int(*sts.Spec.Replicas)), nil
	default:
		return "", fmt.Errorf("scaling not supported for %s", t.gvr)
	}
}

func (w *Workload) makeScaleForm(targets []workloadTarget) (*tview.Form, error) {
	factor := "0"
	if len(targets) == 1 {
		replicas, err := w.currentReplicas(targets[0])
		if err != nil {
			return nil, err
		}
		factor = replicas
	}

	styles := w.App().Styles.Dialog()
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color())

	f.AddInputField("Replicas:", factor, 0, func(textToCheck string, _ rune) bool {
		_, err := strconv.Atoi(textToCheck)
		return err == nil
	}, func(changed string) {
		factor = changed
	})

	f.AddButton("OK", func() {
		defer w.App().Content.RemovePage(scaleDialogKey)
		count, err := strconv.Atoi(factor)
		if err != nil {
			w.App().Flash().Err(err)
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), w.App().Conn().Config().CallTimeout())
		defer cancel()
		for _, t := range targets {
			if err := w.scaleTarget(ctx, t, int32(count)); err != nil {
				slog.Error("Unable to scale resource", slogs.FQN, t.fqn)
				w.App().Flash().Err(err)
				return
			}
		}
		if len(targets) != 1 {
			w.App().Flash().Infof("[%d] %s scaled successfully", len(targets), singularize(w.GVR().R()))
		} else {
			w.App().Flash().Infof("%s %s scaled successfully", targets[0].gvr.R(), targets[0].fqn)
		}
	})
	f.AddButton("Cancel", func() {
		w.App().Content.RemovePage(scaleDialogKey)
	})
	for i := range f.GetButtonCount() {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	return f, nil
}

func (w *Workload) scaleTarget(ctx context.Context, t workloadTarget, replicas int32) error {
	res, err := dao.AccessorFor(w.App().factory, t.gvr)
	if err != nil {
		return err
	}
	scaler, ok := res.(dao.Scalable)
	if !ok {
		return fmt.Errorf("expecting a scalable resource for %q", t.gvr)
	}

	return scaler.Scale(ctx, t.fqn, replicas)
}

func (w *Workload) setImageCmd(evt *tcell.EventKey) *tcell.EventKey {
	t, ok := w.selectedTarget()
	if !ok || !isSupportedWorkloadGVR(t.gvr) {
		return evt
	}

	w.Stop()
	defer w.Start()
	if err := w.showImageDialog(t); err != nil {
		w.App().Flash().Err(err)
	}

	return nil
}

func (w *Workload) showImageDialog(t workloadTarget) error {
	form, err := w.makeSetImageForm(t)
	if err != nil {
		return err
	}
	confirm := tview.NewModalForm("<Set image>", form)
	confirm.SetText(fmt.Sprintf("Set image %s %s", t.gvr, t.fqn))
	confirm.SetDoneFunc(func(int, string) {
		w.App().Content.RemovePage(imageKey)
	})
	w.App().Content.AddPage(imageKey, confirm, false, false)
	w.App().Content.ShowPage(imageKey)

	return nil
}

func (w *Workload) makeSetImageForm(t workloadTarget) (*tview.Form, error) {
	podSpec, err := w.getPodSpec(t)
	if err != nil {
		return nil, err
	}

	lines := make([]*imageFormSpec, 0, len(podSpec.InitContainers)+len(podSpec.Containers))
	for i := range podSpec.InitContainers {
		spec := podSpec.InitContainers[i]
		lines = append(lines, &imageFormSpec{init: true, name: spec.Name, dockerImage: spec.Image})
	}
	for i := range podSpec.Containers {
		spec := podSpec.Containers[i]
		lines = append(lines, &imageFormSpec{name: spec.Name, dockerImage: spec.Image})
	}

	styles := w.App().Styles.Dialog()
	f := tview.NewForm().
		SetItemPadding(0).
		SetButtonsAlign(tview.AlignCenter).
		SetButtonBackgroundColor(styles.ButtonBgColor.Color()).
		SetButtonTextColor(styles.ButtonFgColor.Color()).
		SetLabelColor(styles.LabelFgColor.Color()).
		SetFieldTextColor(styles.FieldFgColor.Color()).
		AddButton("OK", func() {
			defer w.App().Content.RemovePage(imageKey)
			var modified dao.ImageSpecs
			for _, v := range lines {
				if v.modified() {
					modified = append(modified, v.imageSpec())
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), w.App().Conn().Config().CallTimeout())
			defer cancel()
			if err := w.setImages(ctx, t, modified); err != nil {
				slog.Error("Unable to set image name",
					slogs.FQN, t.fqn,
					slogs.Error, err,
				)
				w.App().Flash().Err(err)
				return
			}
			w.App().Flash().Infof("Resource %s:%s image updated successfully", t.gvr, t.fqn)
		}).
		AddButton("Cancel", func() {
			w.App().Content.RemovePage(imageKey)
		})

	for i := range lines {
		ctn := lines[i]
		f.AddInputField(ctn.name, ctn.dockerImage, 0, nil, func(changed string) {
			ctn.newDockerImage = changed
		})
	}
	for i := range f.GetButtonCount() {
		f.GetButton(i).
			SetBackgroundColorActivated(styles.ButtonFocusBgColor.Color()).
			SetLabelColorActivated(styles.ButtonFocusFgColor.Color())
	}

	return f, nil
}

func (w *Workload) getPodSpec(t workloadTarget) (*corev1.PodSpec, error) {
	res, err := dao.AccessorFor(w.App().factory, t.gvr)
	if err != nil {
		return nil, err
	}
	specRes, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return nil, fmt.Errorf("expecting a ContainsPodSpec for %q but got %T", t.gvr, res)
	}

	return specRes.GetPodSpec(t.fqn)
}

func (w *Workload) setImages(ctx context.Context, t workloadTarget, imageSpecs dao.ImageSpecs) error {
	res, err := dao.AccessorFor(w.App().factory, t.gvr)
	if err != nil {
		return err
	}
	specRes, ok := res.(dao.ContainsPodSpec)
	if !ok {
		return fmt.Errorf("expecting a ContainsPodSpec for %q but got %T", t.gvr, res)
	}

	return specRes.SetImages(ctx, t.fqn, imageSpecs)
}

func (w *Workload) portFwdCmd(evt *tcell.EventKey) *tcell.EventKey {
	t, ok := w.selectedTarget()
	if !ok || !isSupportedWorkloadGVR(t.gvr) {
		return evt
	}

	podName, err := w.fetchPodName(t)
	if err != nil {
		w.App().Flash().Err(err)
		return nil
	}
	if err := ensurePodPortFwdAllowed(w.App().factory, podName); err != nil {
		w.App().Flash().Err(err)
		return nil
	}
	if err := showFwdDialog(w, podName, startFwdCB); err != nil {
		w.App().Flash().Err(err)
	}

	return nil
}

func (w *Workload) showPFCmd(evt *tcell.EventKey) *tcell.EventKey {
	t, ok := w.selectedTarget()
	if !ok || !isSupportedWorkloadGVR(t.gvr) {
		return evt
	}

	podName, err := w.fetchPodName(t)
	if err != nil {
		w.App().Flash().Err(err)
		return nil
	}
	if !w.App().factory.Forwarders().IsPodForwarded(podName) {
		w.App().Flash().Errf("no port-forward defined")
		return nil
	}

	pf := NewPortForward(client.PfGVR)
	pf.SetContextFn(w.portForwardContext(t.fqn))
	if err := w.App().inject(pf, false); err != nil {
		w.App().Flash().Err(err)
	}

	return nil
}

func (w *Workload) fetchPodName(t workloadTarget) (string, error) {
	res, err := dao.AccessorFor(w.App().factory, t.gvr)
	if err != nil {
		return "", err
	}
	ctrl, ok := res.(dao.Controller)
	if !ok {
		return "", fmt.Errorf("expecting a controller resource for %q", t.gvr)
	}

	return ctrl.Pod(t.fqn)
}

func (w *Workload) portForwardContext(fqn string) ContextFunc {
	return func(ctx context.Context) context.Context {
		if bc := w.App().BenchFile; bc != "" {
			ctx = context.WithValue(ctx, internal.KeyBenchCfg, bc)
		}

		return context.WithValue(ctx, internal.KeyPath, fqn)
	}
}

func (w *Workload) replicaSetsCmd(evt *tcell.EventKey) *tcell.EventKey {
	t, ok := w.selectedTarget()
	if !ok || t.gvr != client.DpGVR {
		return evt
	}
	dp, err := getDeployment(w.App().factory, t.fqn)
	if err != nil {
		w.App().Flash().Err(err)
		return nil
	}
	showReplicasetsFromSelector(w.App(), t.fqn, dp.Spec.Selector)

	return nil
}
