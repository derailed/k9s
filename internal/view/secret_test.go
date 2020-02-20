package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestSecretNew(t *testing.T) {
	s := view.NewSecret(client.NewGVR("v1/secrets"))

	assert.Nil(t, s.Init(makeCtx()))
	assert.Equal(t, "Secrets", s.Name())
	assert.Equal(t, 6, len(s.Hints()))
}
