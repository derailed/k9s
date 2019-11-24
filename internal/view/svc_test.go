package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestServiceNew(t *testing.T) {
	s := view.NewService("service", "", resource.NewServiceList(nil, ""))
	s.Init(makeCtx())

	assert.Equal(t, "svc", s.Name())
	assert.Equal(t, 22, len(s.Hints()))
}
