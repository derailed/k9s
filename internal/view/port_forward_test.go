package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestPortForwardNew(t *testing.T) {
	po := view.NewPortForward(dao.GVR("forwards"))
	po.Init(makeCtx())

	assert.Equal(t, "PortForwards", po.Name())
	assert.Equal(t, 16, len(po.Hints()))
}
