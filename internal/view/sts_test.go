package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestStatefulSetNew(t *testing.T) {
	s := view.NewStatefulSet("sts", "", resource.NewStatefulSetList(nil, ""))
	s.Init(makeCtx())

	assert.Equal(t, "sts", s.Name())
	assert.Equal(t, 24, len(s.Hints()))
}
