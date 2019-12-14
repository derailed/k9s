package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := view.NewContext(dao.GVR("contexts"))
	ctx.Init(makeCtx())

	assert.Equal(t, "ctx", ctx.Name())
	assert.Equal(t, 10, len(ctx.Hints()))
}
