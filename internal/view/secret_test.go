package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/dao"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestSecretNew(t *testing.T) {
	s := view.NewSecret(dao.GVR("v1/secrets"))
	s.Init(makeCtx())

	assert.Equal(t, "secrets", s.Name())
	assert.Equal(t, 18, len(s.Hints()))
}
