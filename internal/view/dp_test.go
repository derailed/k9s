package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/client"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestDeploy(t *testing.T) {
	v := view.NewDeploy(client.NewGVR("apps/v1/deployments"))

	assert.Nil(t, v.Init(makeCtx()))
	assert.Equal(t, "Deployments", v.Name())
	assert.Equal(t, 14, len(v.Hints()))
}
