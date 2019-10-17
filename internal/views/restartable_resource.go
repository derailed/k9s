package views

import (
	"errors"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
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

	sel := v.masterPage().GetSelectedItem()
	v.stopUpdates()
	defer v.restartUpdates()
	msg := "Please confirm rollout restart for " + sel
	dialog.ShowConfirm(v.Pages, "<Confirm Restart>", msg, func() {
		if err := v.restartRollout(sel); err != nil {
			v.app.Flash().Err(err)
		} else {
			v.app.Flash().Infof("Rollout restart in progress for `%s...", sel)
		}
	}, func() {
		v.showMaster()
	})

	return nil
}

func (v *restartableResourceView) restartRollout(selection string) error {
	r, ok := v.list.Resource().(resource.Restartable)
	if !ok {
		return errors.New("resource is not of type resource.Restartable")
	}
	ns, n := namespaced(selection)

	return r.Restart(ns, n)
}
