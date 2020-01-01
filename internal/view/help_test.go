package view_test

import (
	"strings"
	"testing"

	"github.com/derailed/k9s/internal"
	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestHelp(t *testing.T) {
	ctx := makeCtx()

	app := ctx.Value(internal.KeyApp).(*view.App)
	po := view.NewPod(client.NewGVR("v1/pods"))
	po.Init(ctx)
	app.Content.Push(po)

	v := view.NewHelp()

	assert.Nil(t, v.Init(ctx))
	assert.Equal(t, 17, v.GetRowCount())
	assert.Equal(t, 8, v.GetColumnCount())
	assert.Equal(t, "<ctrl-k>", strings.TrimSpace(v.GetCell(1, 0).Text))
	assert.Equal(t, "Kill", strings.TrimSpace(v.GetCell(1, 1).Text))
}
