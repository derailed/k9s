// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package view

import (
	"context"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/model"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/rs/zerolog/log"
)

// ValueExtender adds values actions to a given viewer.
type ValueExtender struct {
	ResourceViewer
}

// NewValueExtender returns a new extender.
func NewValueExtender(r ResourceViewer) ResourceViewer {
	p := ValueExtender{ResourceViewer: r}
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(func(app *App, model ui.Tabular, gvr client.GVR, path string) {
		p.valuesCmd(nil)
	})

	return &p
}

func (v *ValueExtender) bindKeys(aa *ui.KeyActions) {
	aa.Add(ui.KeyV, ui.NewKeyAction("Values", v.valuesCmd, true))
}

func (v *ValueExtender) valuesCmd(evt *tcell.EventKey) *tcell.EventKey {
	path := v.GetTable().GetSelectedItem()
	if path == "" {
		return evt
	}

	showValues(v.defaultCtx(), v.App(), path, v.GVR())
	return nil
}

func (v *ValueExtender) defaultCtx() context.Context {
	return context.WithValue(context.Background(), internal.KeyFactory, v.App().factory)
}

func showValues(ctx context.Context, app *App, path string, gvr client.GVR) {
	vm := model.NewValues(gvr, path)
	if err := vm.Init(app.factory); err != nil {
		app.Flash().Errf("Initializing the values model failed: %s", err)
	}

	toggleValuesCmd := func(evt *tcell.EventKey) *tcell.EventKey {
		if err := vm.ToggleValues(); err != nil {
			app.Flash().Errf("Values toggle failed: %s", err)
			return nil
		}

		if err := vm.Refresh(ctx); err != nil {
			log.Error().Err(err).Msgf("values refresh failed")
			return nil
		}

		app.Flash().Infof("Values toggled")
		return nil
	}

	v := NewLiveView(app, "Values", vm)
	v.actions.Add(ui.KeyV, ui.NewKeyAction("Toggle All Values", toggleValuesCmd, true))
	if err := v.app.inject(v, false); err != nil {
		v.app.Flash().Err(err)
	}
}
