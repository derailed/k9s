package views

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubernetes/pkg/kubectl"
)

type replicaSetView struct {
	*resourceView
}

func newReplicaSetView(t string, app *appView, list resource.List) resourceViewer {
	v := replicaSetView{newResourceView(t, app, list).(*resourceView)}
	{
		v.extraActionsFn = v.extraActions
		v.switchPage("rs")
	}

	return &v
}

func (v *replicaSetView) extraActions(aa keyActions) {
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
	aa[tcell.KeyCtrlB] = newKeyAction("Rollback", v.rollbackCmd, true)
	aa[tcell.KeyEnter] = newKeyAction("View Pods", v.showPodsCmd, true)
}

func (v *replicaSetView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *replicaSetView) showPodsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	ns, n := namespaced(v.selectedItem)
	rset := k8s.NewReplicaSet(v.app.conn())
	r, err := rset.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching ReplicaSet %s", v.selectedItem)
		v.app.flash(flashErr, err.Error())
		return evt
	}
	rs := r.(*v1.ReplicaSet)

	sel, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		log.Error().Err(err).Msgf("Converting selector for ReplicaSet %s", v.selectedItem)
		v.app.flash(flashErr, err.Error())
		return evt
	}
	showPods(v.app, "", "ReplicaSet", v.selectedItem, sel.String(), "", v.backCmd)

	return nil
}

func (v *replicaSetView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}

func (v *replicaSetView) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	confirm := v.GetPrimitive("confirm").(*tview.Modal)
	confirm.SetText(fmt.Sprintf("Rollback %s %s?", v.list.GetName(), v.selectedItem))
	confirm.SetDoneFunc(func(_ int, button string) {
		if button == "OK" {
			v.app.flash(flashInfo, fmt.Sprintf("Rolling back %s %s", v.list.GetName(), v.selectedItem))
			rollback(v.app, v.selectedItem)
			v.refresh()
		}
		v.switchPage(v.list.GetName())
	})
	v.SwitchToPage("confirm")

	return nil
}

func rollback(app *appView, selectedItem string) bool {
	ns, n := namespaced(selectedItem)
	rset := k8s.NewReplicaSet(app.conn())
	r, err := rset.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching ReplicaSet %s", selectedItem)
		app.flash(flashErr, err.Error())
		return false
	}
	rs := r.(*v1.ReplicaSet)

	var ctrlName, ctrlKind, ctrlAPI string
	for _, ref := range rs.ObjectMeta.OwnerReferences {
		if ref.Controller != nil && *ref.Controller {
			ctrlAPI, ctrlKind, ctrlName = ref.APIVersion, ref.Kind, ref.Name
			break
		}
	}
	if ctrlName == "" || ctrlKind == "" || ctrlAPI == "" {
		app.flash(flashErr, "Unable to find controller for ReplicaSet %s", selectedItem)
		return false
	}

	revision := rs.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	if rs.Status.Replicas != 0 {
		app.flash(flashWarn, "Can not rollback the current replica!")
		return false
	}

	dpr := k8s.NewDeployment(app.conn())
	dep, err := dpr.Get(ns, ctrlName)
	if err != nil {
		log.Error().Err(err).Msgf("Fetching Deployment %s", selectedItem)
		app.flash(flashErr, err.Error())
		return false
	}
	dp := dep.(*appsv1.Deployment)

	vers, err := strconv.Atoi(revision)
	if err != nil {
		log.Error().Err(err).Msg("Revision conversion failed")
		return false
	}

	tokens := strings.Split(ctrlAPI, "/")
	group := ctrlAPI
	if len(tokens) == 2 {
		group = tokens[0]
	}
	rb, err := kubectl.RollbackerFor(schema.GroupKind{group, ctrlKind}, app.conn().DialOrDie())
	if err != nil {
		log.Error().Err(err).Msg("No rollbacker")
		return false
	}

	res, err := rb.Rollback(dp, map[string]string{}, int64(vers), false)
	if err != nil {
		log.Error().Err(err).Msg("Rollback failed")
		return false
	}
	log.Debug().Msgf("Version %s %s", revision, res)
	app.flash(flashInfo, fmt.Sprintf("Version %s %s", revision, res))

	return true
}
