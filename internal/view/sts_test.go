package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestStatefulSetNew(t *testing.T) {
	s := view.NewStatefulSet(dao.GVR("apps/v1/statefulsets"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "StatefulSets", s.Name())
	assert.Equal(t, 16, len(s.Hints()))
}
