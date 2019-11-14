package view

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

// ReplicaSet presents a replicaset viewer.
type ReplicaSet struct {
	*Resource
}

// NewReplicaSet returns a new viewer.
func NewReplicaSet(title, gvr string, list resource.List) ResourceViewer {
	r := ReplicaSet{
		Resource: NewResource(title, gvr, list),
	}
	r.extraActionsFn = r.extraActions
	r.enterFn = r.showPods

	return &r
}

func (r *ReplicaSet) extraActions(aa ui.KeyActions) {
	aa[ui.KeyShiftD] = ui.NewKeyAction("Sort Desired", r.sortColCmd(1, false), false)
	aa[ui.KeyShiftC] = ui.NewKeyAction("Sort Current", r.sortColCmd(2, false), false)
	aa[tcell.KeyCtrlB] = ui.NewKeyAction("Rollback", r.rollbackCmd, true)
}

func (r *ReplicaSet) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	s, err := k8s.NewReplicaSet(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Errf("Replicaset failed %s", err)
	}

	rs := s.(*v1.ReplicaSet)
	l, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		app.Flash().Errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "", r.backCmd)
}

func (r *ReplicaSet) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	r.app.inject(r)

	return nil
}

func (r *ReplicaSet) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	sel := r.masterPage().GetSelectedItem()
	r.showModal(fmt.Sprintf("Rollback %s %s?", r.list.GetName(), sel), func(_ int, button string) {
		if button == "OK" {
			r.app.Flash().Infof("Rolling back %s %s", r.list.GetName(), sel)
			if res, err := rollback(r.app.Conn(), sel); err != nil {
				r.app.Flash().Err(err)
			} else {
				r.app.Flash().Info(res)
			}
			r.refresh()
		}
		r.dismissModal()
	})

	return nil
}

func (r *ReplicaSet) dismissModal() {
	r.Pop()
}

func (r *ReplicaSet) showModal(msg string, done func(int, string)) {
	confirm := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(done)
	r.AddPage("confirm", confirm, false, false)
	r.ShowPage("confirm")
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
	rb, err := polymorphichelpers.RollbackerFor(schema.GroupKind{Group: apiGroup, Kind: kind}, Conn.DialOrDie())
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
