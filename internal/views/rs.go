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
	v.extraActionsFn = v.extraActions
	v.enterFn = v.showPods

	return &v
}

func (v *replicaSetView) extraActions(aa keyActions) {
	aa[KeyShiftD] = newKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[KeyShiftC] = newKeyAction("Sort Current", v.sortColCmd(3, false), true)
	aa[tcell.KeyCtrlB] = newKeyAction("Rollback", v.rollbackCmd, true)
}

func (v *replicaSetView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *replicaSetView) showPods(app *appView, ns, res, sel string) {
	ns, n := namespaced(sel)
	rset := k8s.NewReplicaSet(app.conn())
	r, err := rset.Get(ns, n)
	if err != nil {
		app.flash().errf("Replicaset failed %s", err)
	}

	rs := r.(*v1.ReplicaSet)
	l, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		app.flash().errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}

func (v *replicaSetView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}

func (v *replicaSetView) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	v.showModal(fmt.Sprintf("Rollback %s %s?", v.list.GetName(), v.selectedItem), func(_ int, button string) {
		if button == "OK" {
			v.app.flash().infof("Rolling back %s %s", v.list.GetName(), v.selectedItem)
			rollback(v.app, v.selectedItem)
			v.refresh()
		}
		v.dismissModal()
	})

	return nil
}

func (v *replicaSetView) dismissModal() {
	v.RemovePage("confirm")
	v.switchPage("master")
}

func (v *replicaSetView) showModal(msg string, done func(int, string)) {
	confirm := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(done)
	v.AddPage("confirm", confirm, false, false)
	v.ShowPage("confirm")
}

// ----------------------------------------------------------------------------
// Helpers...

func rollback(app *appView, selectedItem string) bool {
	ns, n := namespaced(selectedItem)
	rset := k8s.NewReplicaSet(app.conn())
	r, err := rset.Get(ns, n)
	if err != nil {
		app.flash().errf("Failed retrieving replicaset %s", err)
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
		app.flash().errf("Unable to find controller for ReplicaSet %s", selectedItem)
		return false
	}

	revision := rs.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	if rs.Status.Replicas != 0 {
		app.flash().warn("Can not rollback the current replica!")
		return false
	}

	dpr := k8s.NewDeployment(app.conn())
	dep, err := dpr.Get(ns, ctrlName)
	if err != nil {
		app.flash().errf("Unable to retrieve deployments %s", err)
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
	app.flash().infof("Version %s %s", revision, res)

	return true
}
