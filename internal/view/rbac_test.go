package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestRbacNew(t *testing.T) {
	v := view.NewRbac("fred", view.ClusterRole, "")
	v.Init(makeCtx())

	assert.Equal(t, "Rbac", v.Name())
	assert.Equal(t, 9, len(v.Hints()))
}
