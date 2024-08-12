// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestCustomResourceDefinitionRender(t *testing.T) {
	c := render.CustomResourceDefinition{}
	r := model1.NewRow(2)

	assert.NoError(t, c.Render(load(t, "crd"), "", &r))
	assert.Equal(t, "-/adapters.config.istio.io", r.ID)
	assert.Equal(t, "adapters", r.Fields[0])
	assert.Equal(t, "config.istio.io", r.Fields[1])
}
