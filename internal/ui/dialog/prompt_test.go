// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"context"
	"testing"
	"time"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tcell/v2"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestShowPrompt(t *testing.T) {
	t.Run("waiting done", func(t *testing.T) {
		a := tview.NewApplication()
		p := ui.NewPages()
		a.SetRoot(p, false)

		ShowPrompt(config.Dialog{}, p, "Running", "Pod", func(context.Context) {
			time.Sleep(time.Millisecond)
		}, func() {
			t.Errorf("unexpected cancellations")
		})
	})

	t.Run("canceled", func(t *testing.T) {
		a := tview.NewApplication()
		p := ui.NewPages()
		a.SetRoot(p, false)

		go ShowPrompt(config.Dialog{}, p, "Running", "Pod", func(ctx context.Context) {
			select {
			case <-time.After(time.Second):
				t.Errorf("expected cancellations")
			case <-ctx.Done():
			}
		}, func() {})

		time.Sleep(time.Second / 2)
		d := p.GetPrimitive(dialogKey).(*tview.ModalForm)
		if assert.NotNil(t, d) {
			d.InputHandler()(tcell.NewEventKey(tcell.KeyEnter, '\n', 0), func(tview.Primitive) {})
		}
	})
}
