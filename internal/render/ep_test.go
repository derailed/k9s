// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestEndpointsRender(t *testing.T) {
	c := render.Endpoints{}
	r := model1.NewRow(4)

	assert.NoError(t, c.Render(load(t, "ep"), "", &r))
	assert.Equal(t, "default/dictionary1", r.ID)
	assert.Equal(t, model1.Fields{"default", "dictionary1", "<none>"}, r.Fields[:3])
}
