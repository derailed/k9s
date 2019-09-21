package views

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/polymorphichelpers"
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

func (v *replicaSetView) extraActions(aa ui.KeyActions) {
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", v.sortColCmd(2, false), true)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", v.sortColCmd(3, false), true)
	aa[tcell.KeyCtrlB] = ui.NewKeyAction("Rollback", v.rollbackCmd, true)
}

func (v *replicaSetView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.masterPage()
		t.SetSortCol(t.NameColIndex()+col, 0, asc)
		t.Refresh()

		return nil
	}
}

func (v *replicaSetView) showPods(app *appView, ns, res, sel string) {
	ns, n := namespaced(sel)
	rset := k8s.NewReplicaSet(app.Conn())
	r, err := rset.Get(ns, n)
	if err != nil {
		app.Flash().Errf("Replicaset failed %s", err)
	}

	rs := r.(*v1.ReplicaSet)
	l, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		app.Flash().Errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "", v.backCmd)
}

func (v *replicaSetView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	v.app.inject(v)

	return nil
}

func (v *replicaSetView) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	sel := v.masterPage().GetSelectedItem()
	v.showModal(fmt.Sprintf("Rollback %s %s?", v.list.GetName(), sel), func(_ int, button string) {
		if button == "OK" {
			v.app.Flash().Infof("Rolling back %s %s", v.list.GetName(), sel)
			if res, err := rollback(v.app.Conn(), sel); err != nil {
				v.app.Flash().Err(err)
			} else {
				v.app.Flash().Info(res)
			}
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

func findRS(Conn k8s.Connection, ns, n string) (*v1.ReplicaSet, error) {
	rset := k8s.NewReplicaSet(Conn)
	r, err := rset.Get(ns, n)
	if err != nil {
		return nil, err
	}
	return r.(*v1.ReplicaSet), nil
}

func findDP(Conn k8s.Connection, ns, n string) (*appsv1.Deployment, error) {
	dp, err := k8s.NewDeployment(Conn).Get(ns, n)
	if err != nil {
		return nil, err
	}
	return dp.(*appsv1.Deployment), nil
}

func controllerInfo(rs *v1.ReplicaSet) (string, string, string, error) {
	for _, ref := range rs.ObjectMeta.OwnerReferences {
		if ref.Controller == nil {
			continue
		}
		log.Debug().Msgf("Controller name %s", ref.Name)
		tokens := strings.Split(ref.APIVersion, "/")
		apiGroup := ref.APIVersion
		if len(tokens) == 2 {
			apiGroup = tokens[0]
		}
		return ref.Name, ref.Kind, apiGroup, nil
	}
	return "", "", "", fmt.Errorf("Unable to find controller for ReplicaSet %s", rs.ObjectMeta.Name)
}

func getRevision(rs *v1.ReplicaSet) (int64, error) {
	revision := rs.ObjectMeta.Annotations["deployment.kubernetes.io/revision"]
	if rs.Status.Replicas != 0 {
		return 0, errors.New("Can not rollback current replica")
	}
	vers, err := strconv.Atoi(revision)
	if err != nil {
		return 0, errors.New("Revision conversion failed")
	}
	return int64(vers), nil
}

func rollback(Conn k8s.Connection, selectedItem string) (string, error) {
	ns, n := namespaced(selectedItem)

	rs, err := findRS(Conn, ns, n)
	if err != nil {
		return "", err
	}
	version, err := getRevision(rs)
	if err != nil {
		return "", err
	}

	name, kind, apiGroup, err := controllerInfo(rs)
	if err != nil {
		return "", err
	}
	rb, err := polymorphichelpers.RollbackerFor(schema.GroupKind{apiGroup, kind}, Conn.DialOrDie())
	if err != nil {
		return "", err
	}
	dp, err := findDP(Conn, ns, name)
	if err != nil {
		return "", err
	}
	res, err := rb.Rollback(dp, map[string]string{}, version, false)
	if err != nil {
		return "", err
	}

	return res, nil
}
