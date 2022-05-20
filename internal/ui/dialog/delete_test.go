package dialog

import (
	"testing"

	"github.com/derailed/k9s/internal/config"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/tview"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestDeleteDialog(t *testing.T) {
	p := ui.NewPages()

	okFunc := func(p *metav1.DeletionPropagation, f bool) {
		assert.Equal(t, propagationOptions[defaultPropagationIdx], p)
		assert.True(t, f)
	}
	caFunc := func() {
		assert.True(t, true)
	}
	ShowDelete(config.Dialog{}, p, "Yo", okFunc, caFunc)

	d := p.GetPrimitive(deleteKey).(*tview.ModalForm)
	assert.NotNil(t, d)

	dismissDelete(p)
	assert.Nil(t, p.GetPrimitive(deleteKey))
}
