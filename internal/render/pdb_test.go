// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of K9s

package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/model1"
	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPodDisruptionBudgetRender(t *testing.T) {
	c := render.PodDisruptionBudget{}
	r := model1.NewRow(9)

	assert.NoError(t, c.Render(load(t, "pdb"), "", &r))
	assert.Equal(t, "default/fred", r.ID)
	assert.Equal(t, model1.Fields{"default", "fred", "2", render.NAValue, "0", "0", "2", "0"}, r.Fields[:8])
}
