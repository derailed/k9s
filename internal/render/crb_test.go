// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestClusterRoleBindingRender(t *testing.T) {
	c := render.ClusterRoleBinding{}
	r := model1.NewRow(5)

	assert.NoError(t, c.Render(load(t, "crb"), "-", &r))
	assert.Equal(t, "-/blee", r.ID)
	assert.Equal(t, model1.Fields{"blee", "blee", "User", "fernand"}, r.Fields[:4])
}
