package view

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/render"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/watch"
	"github.com/derailed/tview"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/kubectl/pkg/polymorphichelpers"
)

// ReplicaSet presents a replicaset viewer.
type ReplicaSet struct {
	ResourceViewer
}

// NewReplicaSet returns a new viewer.
func NewReplicaSet(gvr client.GVR) ResourceViewer {
	r := ReplicaSet{
		ResourceViewer: NewBrowser(gvr),
	}
	r.bindKeys()
	r.GetTable().SetEnterFn(r.showPods)
	r.GetTable().SetColorerFn(render.ReplicaSet{}.ColorerFunc())

	return &r
}

func (r *ReplicaSet) bindKeys() {
	r.Actions().Add(ui.KeyActions{
		ui.KeyShiftD:   ui.NewKeyAction("Sort Desired", r.GetTable().SortColCmd(1, true), false),
		ui.KeyShiftC:   ui.NewKeyAction("Sort Current", r.GetTable().SortColCmd(2, true), false),
		tcell.KeyCtrlB: ui.NewKeyAction("Rollback", r.rollbackCmd, true),
	})
}

func (r *ReplicaSet) showPods(app *App, _, gvr, path string) {
	o, err := app.factory.Get(r.GVR(), path, labels.Everything())
	if err != nil {
		app.Flash().Err(err)
		return
	}

	var rs appsv1.ReplicaSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rs)
	if err != nil {
		app.Flash().Err(err)
	}

	showPodsFromSelector(app, path, rs.Spec.Selector)
}

func (r *ReplicaSet) rollbackCmd(evt *tcell.EventKey) *tcell.EventKey {
	sel := r.GetTable().GetSelectedItem()
	if sel == "" {
		return evt
	}

	r.showModal(fmt.Sprintf("Rollback %s %s?", r.GVR(), sel), func(_ int, button string) {
		if button == "OK" {
			r.App().Flash().Infof("Rolling back %s %s", r.GVR(), sel)
			if res, err := rollback(r.App().factory, sel); err != nil {
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

func findRS(f *watch.Factory, path string) (*v1.ReplicaSet, error) {
	o, err := f.Get("apps/v1/replicasets", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var rs appsv1.ReplicaSet
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &rs)
	if err != nil {
		return nil, err
	}

	return &rs, nil
}

func findDP(f *watch.Factory, path string) (*appsv1.Deployment, error) {
	o, err := f.Get("apps/v1/deployments", path, labels.Everything())
	if err != nil {
		return nil, err
	}

	var dp appsv1.Deployment
	err = runtime.DefaultUnstructuredConverter.FromUnstructured(o.(*unstructured.Unstructured).Object, &dp)
	if err != nil {
		return nil, err
	}

	return &dp, nil
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
		return 0, errors.New("can not rollback current replica")
	}
	vers, err := strconv.Atoi(revision)
	if err != nil {
		return 0, errors.New("revision conversion failed")
	}

	return int64(vers), nil
}

func rollback(f *watch.Factory, path string) (string, error) {
	rs, err := findRS(f, path)
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
	rb, err := polymorphichelpers.RollbackerFor(schema.GroupKind{Group: apiGroup, Kind: kind}, f.Client().DialOrDie())
	if err != nil {
		return "", err
	}
	dp, err := findDP(f, client.FQN(rs.Namespace, name))
	if err != nil {
		return "", err
	}
	res, err := rb.Rollback(dp, map[string]string{}, version, false)
	if err != nil {
		return "", err
	}

	return res, nil
}
