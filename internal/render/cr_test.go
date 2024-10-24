// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestClusterRoleRender(t *testing.T) {
	c := render.ClusterRole{}
	r := model1.NewRow(2)

	assert.NoError(t, c.Render(load(t, "cr"), "-", &r))
	assert.Equal(t, "-/blee", r.ID)
	assert.Equal(t, model1.Fields{"blee"}, r.Fields[:1])
}
