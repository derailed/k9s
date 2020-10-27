package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestPodDisruptionBudgetRender(t *testing.T) {
	c := render.PodDisruptionBudget{}
	r := render.NewRow(9)
	c.Render(load(t, "pdb"), "", &r)

	assert.Equal(t, "default/fred", r.ID)
	assert.Equal(t, render.Fields{"default", "fred", "2", render.NAValue, "0", "0", "2", "0"}, r.Fields[:8])
}
