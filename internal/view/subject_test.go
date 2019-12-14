package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestSubjectNew(t *testing.T) {
	s := view.NewSubject(client.GVR("subjects"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "subjects", s.Name())
	assert.Equal(t, 9, len(s.Hints()))
}
