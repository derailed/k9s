package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestScreenDumpNew(t *testing.T) {
	po := view.NewScreenDump("fred", "blee", nil)
	po.Init(makeCtx())

	assert.Equal(t, "Screen Dumps", po.Name())
	assert.Equal(t, 13, len(po.Hints()))
}
