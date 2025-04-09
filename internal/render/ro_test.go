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

func TestRoleRender(t *testing.T) {
	c := render.Role{}
	r := model1.NewRow(3)

	require.NoError(t, c.Render(load(t, "ro"), "", &r))
	assert.Equal(t, "default/blee", r.ID)
	assert.Equal(t, model1.Fields{"default", "blee"}, r.Fields[:2])
}
