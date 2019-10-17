package views

import (
	"errors"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/gdamore/tcell"
)

type (
	restartableResourceView struct {
		*resourceView
	}
)

func newRestartableResourceViewForParent(parent *resourceView) *restartableResourceView {
	v := restartableResourceView{
		parent,
	}
	parent.extraActionsFn = v.extraActions
	return &v
}

func (v *restartableResourceView) extraActions(aa ui.KeyActions) {
	aa[tcell.KeyCtrlT] = ui.NewKeyAction("Restart Rollout", v.restartCmd, true)
}

func (v *restartableResourceView) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !v.masterPage().RowSelected() {
		return evt
	}

	v.restartRollout(v.masterPage().GetSelectedItem())
	return nil
}

func (v *restartableResourceView) restartRollout(selection string) {
	ns, n := namespaced(selection)

	r, ok := v.list.Resource().(resource.Restartable)
	if !ok {
		v.app.Flash().Err(errors.New("resource is not of type resource.Restartable"))
		return
	}

	err := r.Restart(ns, n)
	if err != nil {
		v.app.Flash().Err(err)
	}
	v.app.Flash().Info("Restarted Rollout")
}
