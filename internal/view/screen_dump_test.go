package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestScreenDumpNew(t *testing.T) {
	po := view.NewScreenDump(dao.GVR("screendumps"))
	po.Init(makeCtx())

	assert.Equal(t, "Screen Dumps", po.Name())
	assert.Equal(t, 12, len(po.Hints()))
}
