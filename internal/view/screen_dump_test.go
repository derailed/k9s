package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestScreenDumpNew(t *testing.T) {
	po := view.NewScreenDump(client.NewGVR("screendumps"))

	assert.Nil(t, po.Init(makeCtx()))
	assert.Equal(t, "ScreenDumps", po.Name())
	assert.Equal(t, 8, len(po.Hints()))
}
