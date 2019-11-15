package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestDaemonSet(t *testing.T) {
	v := view.NewDaemonSet("blee", "", resource.NewDaemonSetList(nil, ""))
	v.Init(makeCtx())

	assert.Equal(t, "ds", v.Name())
	assert.Equal(t, 23, len(v.Hints()))
}
