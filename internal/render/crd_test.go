// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestCustomResourceDefinitionRender(t *testing.T) {
	c := render.CustomResourceDefinition{}
	r := render.NewRow(2)

	assert.NoError(t, c.Render(load(t, "crd"), "", &r))
	assert.Equal(t, "-/adapters.config.istio.io", r.ID)
	assert.Equal(t, render.Fields{"adapters.config.istio.io"}, r.Fields[:1])
}
