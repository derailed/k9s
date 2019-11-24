package view

import (
	"errors"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
)

// RestartExtender represents a restartable resource.
type RestartExtender struct {
	ResourceViewer
}

// NewRestartExtender returns a new extender.
func NewRestartExtender(r ResourceViewer) ResourceViewer {
	re := RestartExtender{ResourceViewer: r}
	re.BindKeys()

	return &re
}

// BindKeys creates additional menu actions.
func (r *RestartExtender) BindKeys() {
	r.Actions().Add(ui.KeyActions{
		tcell.KeyCtrlT: ui.NewKeyAction("Restart", r.restartCmd, true),
	})
}

func (r *RestartExtender) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := r.GetTable().GetSelectedItem()
	if path == "" {
		return nil
	}

	r.Stop()
	defer r.Start()
	msg := "Please confirm rollout restart for " + path
	dialog.ShowConfirm(r.App().Content.Pages, "<Confirm Restart>", msg, func() {
		if err := r.restartRollout(path); err != nil {
			r.App().Flash().Err(err)
		} else {
			r.App().Flash().Infof("Rollout restart in progress for `%s...", path)
		}
	}, func() {})

	return nil
}

func (r *RestartExtender) restartRollout(path string) error {
	s, ok := r.List().Resource().(resource.Restartable)
	if !ok {
		return errors.New("resource is not restartable")
	}
	ns, n := namespaced(path)

	return s.Restart(ns, n)
}
