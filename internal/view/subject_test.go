package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestSubjectNew(t *testing.T) {
	s := view.NewSubject("subject", "", nil)
	s.Init(makeCtx())

	assert.Equal(t, "subject", s.Name())
	assert.Equal(t, 11, len(s.Hints()))
}
