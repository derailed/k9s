// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package dialog

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestErrorDialog(t *testing.T) {
	p := ui.NewPages()

	ShowError(config.Dialog{}, p, "Yo")

	d := p.GetPrimitive(dialogKey).(*tview.ModalForm)
	assert.NotNil(t, d)
	dismiss(p)
	assert.Nil(t, p.GetPrimitive(dialogKey))
}
