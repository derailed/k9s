// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/ui/dialog"
	"github.com/derailed/tcell/v2"
)

// RestartExtender represents a restartable resource.
type RestartExtender struct {
	ResourceViewer
}

// NewRestartExtender returns a new extender.
func NewRestartExtender(v ResourceViewer) ResourceViewer {
	r := RestartExtender{ResourceViewer: v}
	v.AddBindKeysFn(r.bindKeys)

	return &r
}

// BindKeys creates additional menu actions.
func (r *RestartExtender) bindKeys(aa *ui.KeyActions) {
	if r.App().Config.K9s.IsReadOnly() {
		return
	}
	aa.Add(ui.KeyR, ui.NewKeyActionWithOpts("Restart", r.restartCmd,
		ui.ActionOpts{
			Visible:   true,
			Dangerous: true,
		},
	))
}

func (r *RestartExtender) restartCmd(evt *tcell.EventKey) *tcell.EventKey {
	paths := r.GetTable().GetSelectedItems()
	if len(paths) == 0 || paths[0] == "" {
		return nil
	}

	r.Stop()
	defer r.Start()
	msg := fmt.Sprintf("Restart %s %s?", singularize(r.GVR().R()), paths[0])
	if len(paths) > 1 {
		msg = fmt.Sprintf("Restart %d %s?", len(paths), r.GVR().R())
	}
	dialog.ShowConfirm(r.App().Styles.Dialog(), r.App().Content.Pages, "Confirm Restart", msg, func() {
		ctx, cancel := context.WithTimeout(context.Background(), r.App().Conn().Config().CallTimeout())
		defer cancel()
		for _, path := range paths {
			if err := r.restartRollout(ctx, path); err != nil {
				r.App().Flash().Err(err)
			} else {
				r.App().Flash().Infof("Restart in progress for `%s...", path)
			}
		}
	}, func() {})

	return nil
}

func (r *RestartExtender) restartRollout(ctx context.Context, path string) error {
	res, err := dao.AccessorFor(r.App().factory, r.GVR())
	if err != nil {
		return err
	}
	s, ok := res.(dao.Restartable)
	if !ok {
		return errors.New("resource is not restartable")
	}

	return s.Restart(ctx, path)
}

// Helpers...

func singularize(s string) string {
	if strings.LastIndex(s, "s") == len(s)-1 {
		return s[:len(s)-1]
	}

	return s
}
