package render_test

import (
	"testing"

	"github.com/derailed/k9s/internal/render"
	"github.com/stretchr/testify/assert"
)

func TestDeploymentRender(t *testing.T) {
	c := render.Deployment{}
	r := render.NewRow(7)
	c.Render(load(t, "dp"), "", &r)

	assert.Equal(t, "icx/icx-db", r.ID)
	assert.Equal(t, render.Fields{"icx", "icx-db", "1", "1", "1", "1"}, r.Fields[:6])
}
