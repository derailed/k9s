// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEndpointsRender(t *testing.T) {
	c := render.Endpoints{}
	r := model1.NewRow(4)

	require.NoError(t, c.Render(load(t, "ep"), "", &r))
	assert.Equal(t, "ns-1/blee", r.ID)
	assert.Equal(t, model1.Fields{"ns-1", "blee", "10.0.0.67:8080"}, r.Fields[:3])
}
