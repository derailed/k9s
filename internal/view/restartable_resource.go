package view

import (
	"errors"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
)

// RestartableResource presents a viewer with restart option.
type RestartableResource struct {
	*Resource
}

func newRestartableResourceForParent(parent *Resource) *RestartableResource {
	r := RestartableResource{Resource: parent}
	parent.extraActionsFn = r.extraActions

	return &r
}

func (r *RestartableResource) extraActions(aa ui.KeyActions) {
	aa[tcell.KeyCtrlT] = ui.NewKeyAction("Restart Rollout", r.restartCmd, true)
}

func (r *RestartableResource) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	if !r.masterPage().RowSelected() {
		return evt
	}

	sel := r.masterPage().GetSelectedItem()
	r.Stop()
	defer r.Start()
	msg := "Please confirm rollout restart for " + sel
	dialog.ShowConfirm(r.Pages, "<Confirm Restart>", msg, func() {
		if err := r.restartRollout(sel); err != nil {
			r.app.Flash().Err(err)
		} else {
			r.app.Flash().Infof("Rollout restart in progress for `%s...", sel)
		}
	}, func() {})

	return nil
}

func (r *RestartableResource) restartRollout(selection string) error {
	s, ok := r.list.Resource().(resource.Restartable)
	if !ok {
		return errors.New("resource is not of type resource.Restartable")
	}
	ns, n := namespaced(selection)

	return s.Restart(ns, n)
}
