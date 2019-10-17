package dialog

import (
	"testing"

	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
)

func TestConfirmDialog(t *testing.T) {
	a := tview.NewApplication()
	p := tview.NewPages()
	a.SetRoot(p, false)

	ackFunc := func() {
		assert.True(t, true)
	}
	caFunc := func() {
		assert.True(t, true)
	}
	ShowConfirm(p, "Blee", "Yo", ackFunc, caFunc)

	d := p.GetPrimitive(confirmKey).(*tview.ModalForm)
	assert.NotNil(t, d)

	dismissConfirm(p)
	assert.Nil(t, p.GetPrimitive(confirmKey))
}
