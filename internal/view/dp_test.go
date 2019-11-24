package view_test

import (
	"testing"

	"github.com/derailed/k9s/internal/resource"
	"github.com/derailed/k9s/internal/view"
	"github.com/stretchr/testify/assert"
)

func TestDeploy(t *testing.T) {
	v := view.NewDeploy("Deploy", "", resource.NewDeploymentList(nil, ""))
	v.Init(makeCtx())

	assert.Equal(t, "deploy", v.Name())
	assert.Equal(t, 23, len(v.Hints()))

}
