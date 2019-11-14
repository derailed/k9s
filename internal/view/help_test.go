package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/ui"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func newNS(n string) v1.Namespace {
	return v1.Namespace{ObjectMeta: metav1.ObjectMeta{
		Name: n,
	}}
}

func TestHelpNew(t *testing.T) {
	ctx := makeCtx()

	app := ctx.Value(ui.KeyApp).(*view.App)
	po := view.NewPod("Pod", "blee", resource.NewPodList(nil, ""))
	po.Init(ctx)
	app.Content.Push(po)

	v := view.NewHelp()
	v.Init(ctx)

	assert.Equal(t, 32, v.GetRowCount())
	assert.Equal(t, 10, v.GetColumnCount())
	assert.Equal(t, "<esc>", v.GetCell(1, 0).Text)
	assert.Equal(t, "Back", v.GetCell(1, 1).Text)
}
