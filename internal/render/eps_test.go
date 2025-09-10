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

func TestEndpointSliceRender(t *testing.T) {
	c := render.EndpointSlice{}
	r := model1.NewRow(4)

	require.NoError(t, c.Render(load(t, "eps"), "", &r))
	assert.Equal(t, "blee/fred", r.ID)
	assert.Equal(t, model1.Fields{"blee", "fred", "IPv4", "4244", "172.20.0.2,172.20.0.3"}, r.Fields[:5])
}
