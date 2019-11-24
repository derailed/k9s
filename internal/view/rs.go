package view

import (
	"context"
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
	ResourceViewer
}

// NewReplicaSet returns a new viewer.
func NewReplicaSet(title, gvr string, list resource.List) ResourceViewer {
	return &ReplicaSet{
		ResourceViewer: NewResource(title, gvr, list),
	}
}

// Init initializes the component.
func (r *ReplicaSet) Init(ctx context.Context) error {
	if err := r.ResourceViewer.Init(ctx); err != nil {
		return err
	}
	r.bindKeys()
	r.GetTable().SetEnterFn(r.showPods)

	return nil
}

func (r *ReplicaSet) bindKeys() {
	r.Actions().Add(ui.KeyActions{
		ui.KeyShiftD:   ui.NewKeyAction("Sort Desired", r.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Current", r.GetTable().SortColCmd(2, true), false),
		tcell.KeyCtrlB: ui.NewKeyAction("Rollback", r.rollbackCmd, true),
	})
}

func (r *ReplicaSet) showPods(app *App, _, res, sel string) {
	ns, n := namespaced(sel)
	s, err := k8s.NewReplicaSet(app.Conn()).Get(ns, n)
	if err != nil {
		app.Flash().Errf("Replicaset failed %s", err)
	}

	rs, ok := s.(*v1.ReplicaSet)
	if !ok {
		log.Fatal().Msg("Expecting a valid rs")
	}
	l, err := metav1.LabelSelectorAsSelector(rs.Spec.Selector)
	if err != nil {
		app.Flash().Errf("Selector failed %s", err)
		return
	}

	showPods(app, ns, l.String(), "")
}

func (r *ReplicaSet) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := r.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	r.showModal(fmt.Sprintf("Rollback %s %s?", r.List().GetName(), sel), func(_ int, button string) {
		if button == "OK" {
			r.App().Flash().Infof("Rolling back %s %s", r.List().GetName(), sel)
			if res, err := rollback(r.App().Conn(), sel); err != nil {
				r.App().Flash().Err(err)
			} else {
				r.App().Flash().Info(res)
			}
			r.Refresh()
		}
		r.dismissModal()
	})

	return nil
}

func (r *ReplicaSet) dismissModal() {
	r.App().Content.RemovePage("confirm")
}

func (r *ReplicaSet) showModal(msg string, done func(int, string)) {
	confirm := tview.NewModal().
		AddButtons([]string{"Cancel", "OK"}).
		SetTextColor(tcell.ColorFuchsia).
		SetText(msg).
		SetDoneFunc(done)
	r.App().Content.AddPage("confirm", confirm, false, false)
	r.App().Content.ShowPage("confirm")
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
