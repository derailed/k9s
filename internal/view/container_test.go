package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestContainerNew(t *testing.T) {
	po := view.NewContainer(client.GVR("containers"))

	assert.Nil(t, po.Init(makeCtx()))
	assert.Equal(t, "Containers", po.Name())
	assert.Equal(t, 17, len(po.Hints()))
}
