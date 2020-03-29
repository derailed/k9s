package view

import (
	"context"
	"errors"
	"fmt"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/gdamore/tcell"
)

// RestartExtender represents a restartable resource.
type RestartExtender struct {
	ResourceViewer
}

// NewRestartExtender returns a new extender.
func NewRestartExtender(v ResourceViewer) ResourceViewer {
	r := RestartExtender{ResourceViewer: v}
	r.bindKeys(v.Actions())

	return &r
}

// BindKeys creates additional menu actions.
func (r *RestartExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		tcell.KeyCtrlT: ui.NewKeyAction("Restart", r.restartCmd, true),
	})
}

func (r *RestartExtender) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	paths := r.GetTable().GetSelectedItems()
	if len(paths) == 0 {
		return nil
	}

	r.Stop()
	defer r.Start()
	msg := fmt.Sprintf("Restart deployment %s?", paths[0])
	if len(paths) > 1 {
		msg = fmt.Sprintf("Restart %d deployments?", len(paths))
	}
	dialog.ShowConfirm(r.App().Content.Pages, "Confirm Restart", msg, func() {
		ctx, cancel := context.WithTimeout(context.Background(), client.CallTimeout)
		defer cancel()
		for _, path := range paths {
			if err := r.restartRollout(ctx, path); err != nil {
				r.App().Flash().Err(err)
			} else {
				r.App().Flash().Infof("Rollout restart in progress for `%s...", path)
			}
		}
	}, func() {})

	return nil
}

func (r *RestartExtender) restartRollout(ctx context.Context, path string) error {
	res, err := dao.AccessorFor(r.App().factory, r.GVR())
	if err != nil {
		return nil
	}
	s, ok := res.(dao.Restartable)
	if !ok {
		return errors.New("resource is not restartable")
	}

	return s.Restart(ctx, path)
}
