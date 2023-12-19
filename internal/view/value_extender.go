package view

import (
	"context"
	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
)

// ValueExtender adds values actions to a given viewer.
type ValueExtender struct {
	ResourceViewer
}

// NewValueExtender returns a new extender.
func NewValueExtender(r ResourceViewer) ResourceViewer {
	p := ValueExtender{ResourceViewer: r}
	p.AddBindKeysFn(p.bindKeys)
	p.GetTable().SetEnterFn(func(app *App, model ui.Tabular, gvr, path string) {
		p.valuesCmd(nil)
	})

	return &p
}

func (v *ValueExtender) bindKeys(aa ui.KeyActions) {
	aa.Add(ui.KeyActions{
		ui.KeyV: ui.NewKeyAction("Values", v.valuesCmd, true),
	})
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
