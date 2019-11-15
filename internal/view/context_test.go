package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestContext(t *testing.T) {
	ctx := view.NewContext("ctx", "", resource.NewContextList(nil, "fred"))
	ctx.Init(makeCtx())

	assert.Equal(t, "ctx", ctx.Name())
	assert.Equal(t, 12, len(ctx.Hints()))
}
