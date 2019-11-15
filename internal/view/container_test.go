package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestContainerNew(t *testing.T) {
	po := view.NewContainer("Container", resource.NewContainerList(nil, nil), "fred/blee")
	po.Init(makeCtx())

	assert.Equal(t, "containers", po.Name())
	assert.Equal(t, 21, len(po.Hints()))
}
