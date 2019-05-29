package views

import (
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/k8s"
	"github.com/derailed/k9s/internal/resource"
	"github.com/gdamore/tcell"
	"github.com/rs/zerolog/log"
	v1 "k8s.io/api/core/v1"
)

type svcView struct {
	*resourceView
}

func newSvcView(t string, app *appView, list resource.List) resourceViewer {
	v := svcView{newResourceView(t, app, list).(*resourceView)}
	v.extraActionsFn = v.extraActions

	return &v
}

func (v *svcView) extraActions(aa keyActions) {
	aa[KeyShiftT] = newKeyAction("Sort Type", v.sortColCmd(1, false), true)
	aa[tcell.KeyEnter] = newKeyAction("View Pods", v.showPodsCmd, true)
}

func (v *svcView) sortColCmd(col int, asc bool) func(evt *tcell.EventKey) *tcell.EventKey {
	return func(evt *tcell.EventKey) *tcell.EventKey {
		t := v.getTV()
		t.sortCol.index, t.sortCol.asc = t.nameColIndex()+col, asc
		t.refresh()

		return nil
	}
}

func (v *svcView) showPodsCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.rowSelected() {
		return evt
	}

	s := k8s.NewService(v.app.conn())
	ns, n := namespaced(v.selectedItem)
	res, err := s.Get(ns, n)
	if err != nil {
		log.Error().Err(err).Msgf("Fetch service %s", v.selectedItem)
		return nil
	}
	if svc, ok := res.(*v1.Service); ok {
		v.showSvcPods(ns, svc.Spec.Selector, v.backCmd)
	}

	return nil
}

func (v *svcView) backCmd(evt *tcell.EventKey) *tcell.EventKey {
	// Reset namespace to what it was
	v.app.config.SetActiveNamespace(v.list.GetNamespace())
	v.app.inject(v)

	return nil
}

func (v *svcView) showSvcPods(ns string, sel map[string]string, b actionHandler) {
	var s []string
	for k, v := range sel {
		s = append(s, fmt.Sprintf("%s=%s", k, v))
	}
	list := resource.NewPodList(v.app.conn(), ns)
	list.SetLabelSelector(strings.Join(s, ","))

	pv := newPodView("Pods", v.app, list)
	pv.setColorerFn(podColorer)
	pv.setExtraActionsFn(func(aa keyActions) {
		aa[tcell.KeyEsc] = newKeyAction("Back", b, true)
	})
	// set active namespace to service ns.
	v.app.config.SetActiveNamespace(ns)
	v.app.inject(pv)
}
