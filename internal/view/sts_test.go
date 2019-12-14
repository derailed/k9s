package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestStatefulSetNew(t *testing.T) {
	s := view.NewStatefulSet(dao.GVR("apps/v1/statefulsets"))
	s.Init(makeCtx())

	assert.Equal(t, "sts", s.Name())
	assert.Equal(t, 23, len(s.Hints()))
}
